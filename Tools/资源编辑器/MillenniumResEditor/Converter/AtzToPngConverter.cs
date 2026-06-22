using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Text;
using System.Text.Json;
using MillenniumResEditor.Model;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor.Converter
{
    /// <summary>
    /// .atz 多层合成渲染为 PNG 序列帧转换器
    ///
    /// 原版角色渲染流程（来自 CharCls.pas）：
    ///   1. 阴影层（0xxx.atz / 1xxx.atz）
    ///   2. 基础层（a0-5xxx.atz / n0-5xxx.atz，身体+基础装备）
    ///   3. 服装层（a6-8xxx.atz / n6-8xxx.atz，服装+头发+头装备）
    ///   4. 武器层（a9xxx.atz / n9xxx.atz）
    ///
    /// 命名规则：
    ///   前缀 a = 女性，n = 男性
    ///   第1位 = 部位编号（0-9）
    ///   第2-3位 = block（每500帧一个block，对应不同.atz文件）
    ///   示例：a000.atz = 女性身体block0，n121.atz = 男性服装block1
    ///
    /// 合成目标：
    ///   将多层 .atz 中对应帧合成到一张 PNG，输出符合当前客户端 manifest 格式
    /// </summary>
    public class AtzToPngConverter
    {
        /// <summary>角色帧最大尺寸（原版 CharMaxSiez = 160）</summary>
        private const int CharMaxSize = 160;

        /// <summary>角色帧中心点（CharMaxSiez / 2）</summary>
        private const int CharMaxSizeHalf = 80;

        /// <summary>PGK 包提取器（用于从 sprite.pgk 加载 .atz）</summary>
        private readonly PgkExtractor _spritePgk;

        /// <summary>PGK 包提取器（用于从 sys.pgk 加载 .atd）</summary>
        private readonly PgkExtractor _sysPgk;

        /// <summary>.atz 文件缓存（文件名 → 解析结果）</summary>
        private readonly Dictionary<string, AtzParser> _atzCache = new();

        /// <summary>.atd 动画缓存（文件名 → 解析结果）</summary>
        private readonly Dictionary<string, AtdParser> _atdCache = new();

        /// <summary>构造函数</summary>
        /// <param name="spritePgk">sprite.pgk 提取器（包含 .atz 文件）</param>
        /// <param name="sysPgk">sys.pgk 提取器（包含 .atd 文件）</param>
        public AtzToPngConverter(PgkExtractor spritePgk, PgkExtractor sysPgk)
        {
            _spritePgk = spritePgk;
            _sysPgk = sysPgk;
        }

        /// <summary>
        /// 从独立 .atz 文件目录构造（不使用 PGK 包）
        /// </summary>
        public AtzToPngConverter(string atzDir, string atdDir)
        {
            AtzDirectory = atzDir;
            AtdDirectory = atdDir;
        }

        /// <summary>.atz 文件所在目录（独立文件模式）</summary>
        public string AtzDirectory { get; set; }

        /// <summary>.atd 文件所在目录（独立文件模式）</summary>
        public string AtdDirectory { get; set; }

        /// <summary>
        /// 加载 .atz 文件（优先从缓存读取）
        /// </summary>
        public AtzParser LoadAtz(string atzFileName)
        {
            if (_atzCache.TryGetValue(atzFileName, out AtzParser cached))
                return cached;

            var parser = new AtzParser();
            try
            {
                if (_spritePgk != null)
                {
                    // 从 PGK 包加载
                    using var ms = _spritePgk.Extract(atzFileName.ToUpperInvariant());
                    if (ms != null)
                        parser.LoadFromStream(ms);
                }
                else if (!string.IsNullOrEmpty(AtzDirectory))
                {
                    // 从独立文件加载
                    string path = Path.Combine(AtzDirectory, atzFileName);
                    if (File.Exists(path))
                        parser.Load(path);
                }
            }
            catch { }

            _atzCache[atzFileName] = parser;
            return parser;
        }

        /// <summary>
        /// 加载 .atd 动画定义文件
        /// </summary>
        public AtdParser LoadAtd(string atdFileName)
        {
            if (_atdCache.TryGetValue(atdFileName, out AtdParser cached))
                return cached;

            var parser = new AtdParser();
            try
            {
                if (_sysPgk != null)
                {
                    using var ms = _sysPgk.Extract(atdFileName.ToUpperInvariant());
                    if (ms != null)
                        parser.LoadFromStream(ms);
                }
                else if (!string.IsNullOrEmpty(AtdDirectory))
                {
                    string path = Path.Combine(AtdDirectory, atdFileName);
                    if (File.Exists(path))
                        parser.Load(path);
                }
            }
            catch { }

            _atdCache[atdFileName] = parser;
            return parser;
        }

        /// <summary>
        /// 合成渲染单帧到 Bitmap
        /// </summary>
        /// <param name="layers">各层的 .atz 解析结果（按绘制顺序：阴影、基础、服装、武器）</param>
        /// <param name="frameIndex">帧索引（对应 .atz 中的第几帧）</param>
        /// <param name="offsetX">X偏移（来自 .atd 的 px）</param>
        /// <param name="offsetY">Y偏移（来自 .atd 的 py）</param>
        public Bitmap CompositeFrame(List<AtzParser> layers, int frameIndex, int offsetX, int offsetY)
        {
            // 创建合成画布（CharMaxSize x CharMaxSize，中心对齐）
            var bmp = new Bitmap(CharMaxSize, CharMaxSize, PixelFormat.Format32bppArgb);
            using (var g = Graphics.FromImage(bmp))
            {
                g.Clear(Color.Transparent);

                foreach (var layer in layers)
                {
                    if (layer == null || frameIndex < 0 || frameIndex >= layer.Frames.Count)
                        continue;

                    var frame = layer.Frames[frameIndex];
                    if (frame.Width == 0 || frame.Height == 0) continue;

                    // 绘制位置 = 帧偏移 + 动画偏移 + 中心点
                    int x = frame.OffsetX + offsetX + CharMaxSizeHalf;
                    int y = frame.OffsetY + offsetY + CharMaxSizeHalf;

                    using var frameBmp = AtzParser.FrameToBitmap(frame);
                    g.DrawImage(frameBmp, x, y);
                }
            }
            return bmp;
        }

        /// <summary>
        /// 导出角色完整动画序列为 PNG
        /// </summary>
        /// <param name="isMale">是否男性（true=n前缀，false=a前缀）</param>
        /// <param name="bodyId">身体ID（rArr[0]）</param>
        /// <param name="equipmentIds">装备ID数组 [0-9]（rArr[2], rArr[4]...rArr[18]）</param>
        /// <param name="atdIndex">.atd 文件索引（0.atd, 1.atd...）</param>
        /// <param name="outputDir">输出目录</param>
        /// <param name="characterName">角色名（male/female）</param>
        public CharacterManifest ExportCharacter(
            bool isMale,
            int bodyId,
            int[] equipmentIds,
            int atdIndex,
            string outputDir,
            string characterName)
        {
            string prefix = isMale ? "n" : "a";
            string charDir = Path.Combine(outputDir, characterName);
            Directory.CreateDirectory(charDir);

            var manifest = new CharacterManifest
            {
                CharacterName = characterName,
                Animations = new Dictionary<string, List<string>>()
            };

            // 加载 .atd 动画定义
            var atd = LoadAtd($"{atdIndex}.atd");
            if (atd == null || atd.Animations.Count == 0)
            {
                throw new InvalidOperationException($"无法加载 {atdIndex}.atd 动画定义文件");
            }

            // 按动作和方向分组导出
            var groups = new Dictionary<string, List<AtdAnimation>>();
            foreach (var anim in atd.Animations)
            {
                string key = $"{anim.Action}_{anim.Direction}";
                if (!groups.ContainsKey(key))
                    groups[key] = new List<AtdAnimation>();
                groups[key].Add(anim);
            }

            foreach (var kv in groups)
            {
                var anims = kv.Value;
                if (anims.Count == 0) continue;

                var firstAnim = anims[0];
                string actionName = MapActionToManifest(firstAnim.Action);
                string dirName = MapDirectionToManifest(firstAnim.Direction);
                if (actionName == null || dirName == null) continue;

                string manifestKey = $"{actionName}_{dirName}";
                var frameFiles = new List<string>();

                for (int f = 0; f < firstAnim.Frames.Count; f++)
                {
                    var frameInfo = firstAnim.Frames[f];
                    int imageIndex = frameInfo.ImageIndex;

                    // 计算 block 和 bidx
                    int block = imageIndex / 500;
                    int bidx = imageIndex % 500;

                    // 加载各层 .atz
                    var layers = new List<AtzParser>();

                    // 阴影层
                    string shadowName = $"{(isMale ? 1 : 0)}{bodyId}{block}.atz";
                    layers.Add(LoadAtz(shadowName));

                    // 基础层（0-5）
                    for (int i = 0; i <= 5; i++)
                    {
                        int eqId = i == 0 ? bodyId : equipmentIds[i];
                        if (i != 0 && eqId == 0) continue;
                        string name = $"{prefix}{i}{eqId}{block}.atz";
                        layers.Add(LoadAtz(name));
                    }

                    // 服装层（6-8）
                    for (int i = 6; i <= 8; i++)
                    {
                        int eqId = equipmentIds[i];
                        if (eqId == 0) continue;
                        string name = $"{prefix}{i}{eqId}{block}.atz";
                        layers.Add(LoadAtz(name));
                    }

                    // 武器层（9）
                    if (equipmentIds[9] != 0)
                    {
                        string name = $"{prefix}9{equipmentIds[9]}{block}.atz";
                        layers.Add(LoadAtz(name));
                    }

                    // 合成渲染
                    using var bmp = CompositeFrame(layers, bidx, frameInfo.OffsetX, frameInfo.OffsetY);

                    // 裁剪到有效区域（可选：移除透明边距）
                    var cropped = CropTransparent(bmp);

                    // 保存 PNG
                    string fileName = $"{characterName}/{actionName}_{dirName}_{f}.png";
                    string filePath = Path.Combine(outputDir, fileName.Replace('/', '\\'));
                    cropped.Save(filePath, ImageFormat.Png);
                    cropped.Dispose();

                    frameFiles.Add(fileName);
                }

                manifest.Animations[manifestKey] = frameFiles;
            }

            return manifest;
        }

        /// <summary>
        /// 裁剪透明边距，减小 PNG 尺寸
        /// </summary>
        private Bitmap CropTransparent(Bitmap src)
        {
            int minX = src.Width, minY = src.Height, maxX = 0, maxY = 0;
            bool hasContent = false;

            var rect = new Rectangle(0, 0, src.Width, src.Height);
            var bmpData = src.LockBits(rect, ImageLockMode.ReadOnly, PixelFormat.Format32bppArgb);

            try
            {
                unsafe
                {
                    byte* scan0 = (byte*)bmpData.Scan0;
                    int stride = bmpData.Stride;

                    for (int y = 0; y < src.Height; y++)
                    {
                        byte* row = scan0 + y * stride;
                        for (int x = 0; x < src.Width; x++)
                        {
                            if (row[x * 4 + 3] > 0) // Alpha > 0
                            {
                                if (x < minX) minX = x;
                                if (x > maxX) maxX = x;
                                if (y < minY) minY = y;
                                if (y > maxY) maxY = y;
                                hasContent = true;
                            }
                        }
                    }
                }
            }
            finally
            {
                src.UnlockBits(bmpData);
            }

            if (!hasContent) return src;

            int w = maxX - minX + 1;
            int h = maxY - minY + 1;
            var result = new Bitmap(w, h, PixelFormat.Format32bppArgb);
            using (var g = Graphics.FromImage(result))
            {
                g.DrawImage(src, new Rectangle(0, 0, w, h), new Rectangle(minX, minY, w, h), GraphicsUnit.Pixel);
            }
            src.Dispose();
            return result;
        }

        /// <summary>
        /// 将原版动作映射到当前客户端 manifest 动作名
        /// </summary>
        private static string MapActionToManifest(AnimAction action)
        {
            return action switch
            {
                AnimAction.Turn or AnimAction.Turnning => "idle",
                AnimAction.Move => "walk",
                AnimAction.Move1 => "walk", // 谨慎走
                AnimAction.Move4 => "walk", // 跑步
                AnimAction.Hit => "attack",
                AnimAction.Hit1 or AnimAction.Hit2 or AnimAction.Hit3 or AnimAction.Hit4
                or AnimAction.Hit5 or AnimAction.Hit6 or AnimAction.Hit7 or AnimAction.Hit8 or AnimAction.Hit9 => "attack",
                AnimAction.Die => "dead",
                AnimAction.SeatDown or AnimAction.StandUp => "idle",
                _ => null
            };
        }

        /// <summary>
        /// 将原版方向（DR_0~DR_7）映射到当前客户端方向名
        /// 原版：0=上，1=右上，2=右，3=右下，4=下，5=左下，6=左，7=左上
        /// 当前：down, down_left, left, up_left, up, up_right, right, down_right
        /// </summary>
        private static string MapDirectionToManifest(int direction)
        {
            return direction switch
            {
                0 => "up",
                1 => "up_right",
                2 => "right",
                3 => "down_right",
                4 => "down",
                5 => "down_left",
                6 => "left",
                7 => "up_left",
                _ => null
            };
        }

        /// <summary>
        /// 生成 manifest.json 清单文件
        /// </summary>
        public static void GenerateManifest(string outputPath, Dictionary<string, CharacterManifest> characters)
        {
            var manifest = new
            {
                version = 2,
                description = "角色资源清单 - 从原版 .atz 转换",
                base_dir = "assets/characters/",
                frame_duration = 120,
                design_size = 48,
                directions = new[] { "down", "down_left", "left", "up_left", "up", "up_right", "right", "down_right" },
                characters = new Dictionary<string, object>()
            };

            foreach (var kv in characters)
            {
                var charAnims = new Dictionary<string, string[]>();
                foreach (var anim in kv.Value.Animations)
                {
                    charAnims[anim.Key] = anim.Value.ToArray();
                }
                manifest.characters[kv.Key] = charAnims;
            }

            var options = new JsonSerializerOptions
            {
                WriteIndented = true,
                Encoder = System.Text.Encodings.Web.JavaScriptEncoder.UnsafeRelaxedJsonEscaping
            };

            string json = JsonSerializer.Serialize(manifest, options);
            File.WriteAllText(outputPath, json, new UTF8Encoding(false));
        }
    }

    /// <summary>
    /// 角色导出清单（中间结构，用于生成 manifest.json）
    /// </summary>
    public class CharacterManifest
    {
        public string CharacterName { get; set; }
        public Dictionary<string, List<string>> Animations { get; set; }
    }
}
