using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor.Converter
{
    /// <summary>
    /// 批量角色库导出器
    ///
    /// 功能：
    ///   1. 扫描 PGK 包或目录中的所有 .atz/.atd 文件
    ///   2. 按命名规则自动识别角色（前缀+部位ID）
    ///   3. 批量导出所有角色的所有动作/方向为 PNG 序列
    ///   4. 生成完整的 manifest.json 清单文件
    ///   5. 支持增量更新（只导出变化的文件）
    ///
    /// 输出结构：
    ///   outputDir/
    ///   ├── manifest.json          # 全局清单
    ///   ├── characters/
    ///   │   ├── male_001/         # 男性角色1
    ///   │   │   ├── idle_up_0.png
    ///   │   │   ├── idle_up_1.png
    ///   │   │   ├── walk_right_0.png
    ///   │   │   └── ...
    ///   │   ├── female_002/       # 女性角色2
    ///   │   └── ...
    ///   └── _meta/
    ///       ├── role_index.json    # 角色索引（ID→名称映射）
    ///       └── export_log.txt     # 导出日志
    /// </summary>
    public class BatchCharacterExporter
    {
        private readonly PgkExtractor _spritePgk;
        private readonly PgkExtractor _sysPgk;
        private readonly AtzToPngConverter _converter;
        private readonly Dictionary<string, AtzParser> _atzCache = new();
        private readonly Dictionary<string, AtdParser> _atdCache = new();

        /// <summary>独立 .atz 文件目录（非PGK模式）</summary>
        public string AtzDirectory { get; set; }

        /// <summary>独立 .atd 文件目录</summary>
        public string AtdDirectory { get; set; }

        /// <summary>输出基础目录</summary>
        public string OutputBase { get; set; }

        /// <summary>是否覆盖已存在的文件</summary>
        public bool OverwriteExisting { get; set; } = false;

        /// <summary>导出进度回调</summary>
        public Action<string, int, int> ProgressCallback { get; set; }

        /// <summary>日志记录</summary>
        public List<string> Log { get; } = new List<string>();

        /// <summary>统计信息</summary>
        public ExportStatistics Statistics { get; } = new ExportStatistics();

        /// <summary>从 PGK 包构造</summary>
        public BatchCharacterExporter(PgkExtractor spritePgk, PgkExtractor sysPgk)
        {
            _spritePgk = spritePgk;
            _sysPgk = sysPgk;
            _converter = new AtzToPngConverter(spritePgk, sysPgk);
        }

        /// <summary>从独立文件目录构造</summary>
        public BatchCharacterExporter(string atzDir, string atdDir)
        {
            AtzDirectory = atzDir;
            AtdDirectory = atdDir;
            _converter = new AtzToPngConverter(atzDir, atdDir);
        }

        /// <summary>
        /// 扫描并导出所有角色
        /// </summary>
        /// <param name="maxBodyId">最大身体ID扫描范围（0-maxBodyId）</param>
        /// <param name="maxAtdIndex">最大ATD索引范围（0-maxAtdIndex）</param>
        /// <param name="exportDefaultEquipment">是否导出默认装备组合</param>
        public void ExportAll(int maxBodyId = 200, int maxAtdIndex = 50, bool exportDefaultEquipment = true)
        {
            if (string.IsNullOrEmpty(OutputBase))
                throw new InvalidOperationException("请先设置 OutputBase 输出目录");

            DateTime startTime = DateTime.Now;
            Log.Clear();
            Statistics.Reset();
            Log.Add($"=== 批量角色库导出开始 ===");
            Log.Add($"时间: {startTime:yyyy-MM-dd HH:mm:ss}");
            Log.Add($"模式: {(_spritePgk != null ? "PGK包" : "独立文件")}");
            Log.Add($"输出: {OutputBase}");
            Log.Add($"参数: maxBodyId={maxBodyId}, maxAtdIndex={maxAtdIndex}");

            Directory.CreateDirectory(Path.Combine(OutputBase, "characters"));
            Directory.CreateDirectory(Path.Combine(OutputBase, "_meta"));

            // 1. 扫描所有可用的 ATZ 文件
            var availableAtzFiles = ScanAvailableAtzFiles();
            Log.Add($"发现 {availableAtzFiles.Count} 个 ATZ 文件");

            // 2. 扫描所有可用的 ATD 文件
            var availableAtdFiles = ScanAvailableAtdFiles(maxAtdIndex);
            Log.Add($"发现 {availableAtdFiles.Count} 个 ATD 文件");

            // 3. 按性别和身体ID分组导出
            var allCharacters = new Dictionary<string, CharacterExportInfo>();
            int totalCharacters = (maxBodyId + 1) * 2; // 男+女

            for (int gender = 0; gender <= 1; gender++)
            {
                bool isMale = (gender == 0);
                string prefix = isMale ? "n" : "a";
                string genderName = isMale ? "male" : "female";

                for (int bodyId = 0; bodyId <= maxBodyId; bodyId++)
                {
                    // 检查该角色是否有对应的 ATZ 文件
                    if (!HasCharacterData(prefix, bodyId, availableAtzFiles))
                        continue;

                    string charKey = $"{genderName}_{bodyId:D3}";
                    int exported = Statistics.ExportedCount;
                    ProgressCallback?.Invoke($"正在处理: {charKey}", exported, totalCharacters);

                    try
                    {
                        // 导出默认装备组合（全0，即裸体）
                        int[] defaultEquip = new int[10];
                        var result = ExportSingleCharacter(
                            isMale, bodyId, defaultEquip,
                            0, // 默认使用 0.atd
                            charKey);

                        if (result != null && result.Animations.Count > 0)
                        {
                            var exportInfo = new CharacterExportInfo
                            {
                                CharacterName = result.CharacterName,
                                Animations = result.Animations,
                                DisplayName = charKey,
                                IsMale = isMale,
                                BodyId = bodyId,
                                TotalFrames = result.Animations.Values.Sum(list => list.Count)
                            };
                            allCharacters[charKey] = exportInfo;
                            Statistics.ExportedCount++;
                            Log.Add($"[OK] {charKey}: {result.Animations.Count} 个动画");
                        }
                    }
                    catch (Exception ex)
                    {
                        Statistics.FailedCount++;
                        Log.Add($"[FAIL] {charKey}: {ex.Message}");
                    }
                }
            }

            // 4. 生成全局 manifest.json
            GenerateGlobalManifest(allCharacters);

            // 5. 生成角色索引
            GenerateRoleIndex(allCharacters);

            // 6. 保存日志
            SaveLog();

            TimeSpan elapsed = DateTime.Now - startTime;
            Statistics.ElapsedSeconds = elapsed.TotalSeconds;
            Log.Add($"\n=== 导出完成 ===");
            Log.Add($"耗时: {elapsed:mm\\mss\\s}");
            Log.Add($"成功: {Statistics.ExportedCount} 个角色");
            Log.Add($"失败: {Statistics.FailedCount} 个角色");
            Log.Add($"总帧数: {Statistics.TotalFramesExported}");

            int finalCount = Statistics.ExportedCount;
            ProgressCallback?.Invoke($"导出完成！成功 {finalCount} 个角色",
                finalCount, finalCount);
        }

        /// <summary>
        /// 导出单个角色（完整版，支持自定义装备）
        /// </summary>
        public CharacterManifest ExportSingleCharacter(
            bool isMale,
            int bodyId,
            int[] equipmentIds,
            int atdIndex,
            string characterName)
        {
            return _converter.ExportCharacter(
                isMale, bodyId, equipmentIds, atdIndex,
                Path.Combine(OutputBase, "characters"),
                characterName);
        }

        /// <summary>
        /// 扫描所有可用的 ATZ 文件
        /// </summary>
        private List<string> ScanAvailableAtzFiles()
        {
            var files = new List<string>();

            if (_spritePgk != null)
            {
                foreach (var entry in _spritePgk.Files)
                {
                    if (entry.Name.EndsWith(".ATZ", StringComparison.OrdinalIgnoreCase))
                        files.Add(entry.Name);
                }
            }
            else if (!string.IsNullOrEmpty(AtzDirectory) && Directory.Exists(AtzDirectory))
            {
                files.AddRange(Directory.GetFiles(AtzDirectory, "*.atz", SearchOption.TopDirectoryOnly));
            }

            return files;
        }

        /// <summary>
        /// 扫描所有可用的 ATD 文件
        /// </summary>
        private List<string> ScanAvailableAtdFiles(int maxIndex)
        {
            var files = new List<string>();

            if (_sysPgk != null)
            {
                for (int i = 0; i <= maxIndex; i++)
                {
                    string name = $"{i}.ATD";
                    var entry = _sysPgk.Files.Find(f =>
                        f.Name.Equals(name, StringComparison.OrdinalIgnoreCase));
                    if (entry != null)
                        files.Add(name);
                }
            }
            else if (!string.IsNullOrEmpty(AtdDirectory) && Directory.Exists(AtdDirectory))
            {
                for (int i = 0; i <= maxIndex; i++)
                {
                    string path = Path.Combine(AtdDirectory, $"{i}.atd");
                    if (File.Exists(path))
                        files.Add(path);
                }
            }

            return files;
        }

        /// <summary>
        /// 检查指定角色是否有数据
        /// </summary>
        private bool HasCharacterData(string prefix, int bodyId, List<string> availableFiles)
        {
            // 检查身体 ATZ 文件是否存在（至少 block 0）
            string pattern = $"{prefix}0{bodyId}0.ATZ";

            if (_spritePgk != null)
            {
                return availableFiles.Exists(f =>
                    f.Equals(pattern, StringComparison.OrdinalIgnoreCase));
            }
            else
            {
                string path = Path.Combine(AtzDirectory, pattern.ToLowerInvariant());
                return File.Exists(path);
            }
        }

        /// <summary>
        /// 生成全局 manifest.json
        /// </summary>
        private void GenerateGlobalManifest(Dictionary<string, CharacterExportInfo> characters)
        {
            var manifest = new
            {
                version = 3,
                description = "千年角色资源批量导出 - 完整库",
                export_time = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss"),
                base_dir = "assets/characters/",
                frame_duration = 120,
                design_size = 48,
                directions = new[] { "down", "down_left", "left", "up_left", "up", "up_right", "right", "down_right" },
                total_characters = characters.Count,
                statistics = new
                {
                    exported_count = Statistics.ExportedCount,
                    failed_count = Statistics.FailedCount,
                    total_frames = Statistics.TotalFramesExported
                },
                characters = new Dictionary<string, object>()
            };

            foreach (var kv in characters)
            {
                var anims = new Dictionary<string, string[]>();
                foreach (var anim in kv.Value.Animations)
                {
                    anims[anim.Key] = anim.Value.ToArray();
                }
                manifest.characters[kv.Key] = new
                {
                    name = kv.Value.DisplayName,
                    gender = kv.Value.IsMale ? "male" : "female",
                    body_id = kv.Value.BodyId,
                    animations = anims,
                    frame_count = kv.Value.TotalFrames
                };
            }

            var options = new JsonSerializerOptions
            {
                WriteIndented = true,
                Encoder = System.Text.Encodings.Web.JavaScriptEncoder.UnsafeRelaxedJsonEscaping
            };

            string json = JsonSerializer.Serialize(manifest, options);
            string path = Path.Combine(OutputBase, "manifest.json");
            File.WriteAllText(path, json, new UTF8Encoding(false));
        }

        /// <summary>
        /// 生成角色索引文件（用于快速查找）
        /// </summary>
        private void GenerateRoleIndex(Dictionary<string, CharacterExportInfo> characters)
        {
            var index = new
            {
                version = 1,
                generated_at = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss"),
                total = characters.Count,
                roles = new Dictionary<string, object>()
            };

            foreach (var kv in characters)
            {
                index.roles[kv.Key] = new
                {
                    display_name = kv.Value.DisplayName,
                    file_path = $"characters/{kv.Key}/",
                    animation_count = kv.Value.Animations.Count,
                    total_frames = kv.Value.TotalFrames
                };
            }

            var options = new JsonSerializerOptions { WriteIndented = true };
            string json = JsonSerializer.Serialize(index, options);
            string path = Path.Combine(OutputBase, "_meta", "role_index.json");
            File.WriteAllText(path, json, new UTF8Encoding(false));
        }

        /// <summary>保存导出日志</summary>
        private void SaveLog()
        {
            string path = Path.Combine(OutputBase, "_meta", "export_log.txt");
            File.WriteAllLines(path, Log);
        }
    }

    /// <summary>
    /// 角色导出详细信息（扩展版 CharacterManifest）
    /// </summary>
    public class CharacterExportInfo : CharacterManifest
    {
        public string DisplayName { get; set; }
        public bool IsMale { get; set; }
        public int BodyId { get; set; }
        public int TotalFrames { get; set; }
    }

    /// <summary>
    /// 导出统计信息
    /// </summary>
    public class ExportStatistics
    {
        public int ExportedCount { get; set; }
        public int FailedCount { get; set; }
        public int TotalFramesExported { get; set; }
        public double ElapsedSeconds { get; set; }

        public void Reset()
        {
            ExportedCount = 0;
            FailedCount = 0;
            TotalFramesExported = 0;
            ElapsedSeconds = 0;
        }
    }
}
