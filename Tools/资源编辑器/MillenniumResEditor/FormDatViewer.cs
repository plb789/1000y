using System;
using System.Collections.Generic;
using System.Drawing;
using System.IO;
using System.Linq;
using System.Text;
using System.Windows.Forms;

namespace MillenniumResEditor
{
    /// <summary>
    /// 通用 .DAT 资源包查看器
    /// 支持查看和解包千年游戏的各种 .dat 格式资源文件
    /// </summary>
    public partial class FormDatViewer : Form
    {
        private DatPackage _package;

        public FormDatViewer()
        {
            InitializeComponent();
        }

        // ==================== 事件处理 ====================

        private void btnOpenDat_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog
            {
                Filter = "DAT资源包(*.dat)|*.dat|DAT文件(*.DAT)|*.DAT|所有文件(*.*)|*.*",
                Title = "选择 DAT 资源包文件"
            };

            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                _package = new DatPackage();
                _package.Load(dlg.FileName);

                lstFiles.Items.Clear();
                foreach (var entry in _package.Entries)
                {
                    lstFiles.Items.Add($"{entry.Name} ({FormatSize(entry.Size)})");
                }

                lblInfo.Text = $"文件: {Path.GetFileName(dlg.FileName)} | 类型: {_package.PackageType} | 条目数: {_package.Entries.Count} | 总大小: {FormatSize(_package.TotalSize)}";
                txtPreview.Clear();
                lblFileName.Text = "文件名: -";
                lblFileSize.Text = "大小: - 字节";
                lblOffset.Text = "偏移: -";
            }
            catch (Exception ex)
            {
                MessageBox.Show($"加载失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void btnExtractAll_Click(object sender, EventArgs e)
        {
            if (_package == null || _package.Entries.Count == 0)
            {
                MessageBox.Show("请先打开 DAT 文件", "提示");
                return;
            }

            using var dlg = new FolderBrowserDialog { Description = "选择解包输出目录" };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            string outputDir = dlg.SelectedPath;
            int extracted = 0;
            int failed = 0;

            foreach (var entry in _package.Entries)
            {
                try
                {
                    string filePath = Path.Combine(outputDir, entry.Name);
                    string dir = Path.GetDirectoryName(filePath);
                    if (!Directory.Exists(dir))
                    {
                        Directory.CreateDirectory(dir);
                    }

                    byte[] data = _package.Extract(entry.Name);
                    File.WriteAllBytes(filePath, data);
                    extracted++;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[警告] 解包失败: {entry.Name} - {ex.Message}");
                    failed++;
                }
            }

            MessageBox.Show($"解包完成!\n成功: {extracted}\n失败: {failed}\n\n输出目录: {outputDir}", "解包结果");
        }

        private void btnExtractSelected_Click(object sender, EventArgs e)
        {
            if (_package == null || lstFiles.SelectedIndex < 0)
            {
                MessageBox.Show("请先打开 DAT 文件并选择要解包的条目", "提示");
                return;
            }

            using var dlg = new FolderBrowserDialog { Description = "选择解包输出目录" };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            var selectedIndices = lstFiles.SelectedIndices;
            int extracted = 0;

            foreach (int index in selectedIndices)
            {
                if (index < 0 || index >= _package.Entries.Count) continue;

                var entry = _package.Entries[index];
                try
                {
                    string filePath = Path.Combine(dlg.SelectedPath, entry.Name);
                    string dir = Path.GetDirectoryName(filePath);
                    if (!Directory.Exists(dir))
                    {
                        Directory.CreateDirectory(dir);
                    }

                    byte[] data = _package.Extract(entry.Name);
                    File.WriteAllBytes(filePath, data);
                    extracted++;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[警告] 解包失败: {entry.Name} - {ex.Message}");
                }
            }

            MessageBox.Show($"已解包 {extracted} 个文件到:\n{dlg.SelectedPath}", "解包完成");
        }

        private void btnSearch_Click(object sender, EventArgs e)
        {
            if (_package == null || string.IsNullOrEmpty(txtSearch.Text)) return;

            string search = txtSearch.Text.ToLower();
            int foundIndex = -1;

            for (int i = 0; i < lstFiles.Items.Count; i++)
            {
                if (lstFiles.Items[i].ToString().ToLower().Contains(search))
                {
                    foundIndex = i;
                    break;
                }
            }

            if (foundIndex >= 0)
            {
                lstFiles.SelectedIndex = foundIndex;
            }
            else
            {
                MessageBox.Show($"未找到包含 \"{txtSearch.Text}\" 的文件", "搜索结果");
            }
        }

        private void lstFiles_SelectedIndexChanged(object sender, EventArgs e)
        {
            if (lstFiles.SelectedIndex < 0 || _package == null) return;

            try
            {
                var entry = _package.Entries[lstFiles.SelectedIndex];
                byte[] data = _package.Extract(entry.Name);

                lblFileName.Text = $"文件名: {entry.Name}";
                lblFileSize.Text = $"大小: {data.Length} 字节";
                lblOffset.Text = $"偏移: 0x{entry.Offset:X8}";

                ShowPreview(data);
            }
            catch (Exception ex)
            {
                txtPreview.Text = $"预览失败: {ex.Message}";
            }
        }

        private void chkHexView_CheckedChanged(object sender, EventArgs e)
        {
            if (lstFiles.SelectedIndex >= 0 && _package != null)
            {
                lstFiles_SelectedIndexChanged(sender, e);
            }
        }

        private void btnPreviewMap_Click(object sender, EventArgs e)
        {
            if (lstFiles.SelectedIndex < 0 || _package == null)
            {
                MessageBox.Show("请先选择一个 .bin 文件", "提示");
                return;
            }

            try
            {
                var entry = _package.Entries[lstFiles.SelectedIndex];
                byte[] data = _package.Extract(entry.Name);

                // 尝试作为地图块数据可视化
                RenderMapBlockPreview(data, entry.Name);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"预览失败: {ex.Message}", "错误");
            }
        }

        // ==================== 辅助方法 ====================

        private void ShowPreview(byte[] data)
        {
            if (chkHexView.Checked)
            {
                ShowHexPreview(data);
            }
            else
            {
                ShowTextPreview(data);
            }
        }

        private void ShowHexPreview(byte[] data)
        {
            var sb = new StringBuilder();

            // 显示前 8KB 的十六进制数据（避免界面卡顿）
            int previewLength = Math.Min(data.Length, 8192);

            sb.AppendLine("偏移     00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F  ASCII");
            sb.AppendLine(new string('-', 80));

            for (int offset = 0; offset < previewLength; offset += 16)
            {
                sb.AppendFormat("{0:X8}  ", offset);

                // 十六进制部分
                for (int i = 0; i < 16; i++)
                {
                    if (offset + i < previewLength)
                    {
                        sb.AppendFormat("{0:X2} ", data[offset + i]);
                    }
                    else
                    {
                        sb.Append("   ");
                    }
                }

                sb.Append(" ");

                // ASCII 部分
                for (int i = 0; i < 16 && offset + i < previewLength; i++)
                {
                    char c = (char)data[offset + i];
                    sb.Append(char.IsControl(c) ? '.' : c);
                }

                sb.AppendLine();
            }

            if (data.Length > previewLength)
            {
                sb.AppendLine($"\n... (共 {data.Length} 字节，仅显示前 {previewLength} 字节)");
            }

            txtPreview.Text = sb.ToString();
        }

        private void ShowTextPreview(byte[] data)
        {
            try
            {
                // 尝试以文本方式显示
                string text = Encoding.UTF8.GetString(data);

                // 过滤不可打印字符
                var sb = new StringBuilder();
                foreach (char c in text)
                {
                    if (char.IsControl(c) && c != '\r' && c != '\n' && c != '\t')
                    {
                        sb.Append('.');
                    }
                    else
                    {
                        sb.Append(c);
                    }
                }

                // 限制显示长度
                if (sb.Length > 50000)
                {
                    sb.Remove(50000, sb.Length - 50000);
                    sb.Append("\n... (内容过长，已截断)");
                }

                txtPreview.Text = sb.ToString();
            }
            catch
            {
                txtPreview.Text = "[无法以文本格式显示]";
            }
        }

        private static string FormatSize(long bytes)
        {
            string[] sizes = { "B", "KB", "MB", "GB" };
            double len = bytes;
            int order = 0;

            while (len >= 1024 && order < sizes.Length - 1)
            {
                order++;
                len /= 1024;
            }

            return $"{len:0.##} {sizes[order]}";
        }

        /// <summary>
        /// 渲染地图块数据为可视化图像
        /// TMapCell 结构 (12字节/单元格):
        /// - TileId: word(2B) | TileNumber: byte(1B)
        /// - TileOverId: word(2B) | TileOverNumber: byte(1B)
        /// - ObjectId: word(2B) | ObjectNumber: byte(1B)
        /// - RoofId: word(2B) | boMove: byte(1B)
        /// </summary>
        private void RenderMapBlockPreview(byte[] data, string fileName)
        {
            const int CELL_SIZE = 12; // TMapCell 大小
            int cellCount = data.Length / CELL_SIZE;

            if (cellCount < 4)
            {
                // 数据太小，可能是索引或标志
                lblMapInfo.Text = $"数据过小 ({data.Length}B)，非标准地图块";
                picMapPreview.Image = null;
                return;
            }

            // 计算网格尺寸（尝试方形或接近方形）
            int gridWidth = (int)Math.Ceiling(Math.Sqrt(cellCount));
            int gridHeight = (int)Math.Ceiling((double)cellCount / gridWidth);

            // 限制显示尺寸
            if (gridWidth > 64) gridWidth = 64;
            if (gridHeight > 64) gridHeight = 64;

            const int PIXEL_PER_CELL = 8; // 每个单元格的像素大小
            int bmpWidth = gridWidth * PIXEL_PER_CELL;
            int bmpHeight = gridHeight * PIXEL_PER_CELL;

            using var bmp = new Bitmap(bmpWidth, bmpHeight);
            using var g = Graphics.FromImage(bmp);
            g.Clear(Color.Black);

            // 收集所有 TileId 用于生成颜色映射
            HashSet<ushort> uniqueTiles = new HashSet<ushort>();
            for (int i = 0; i < cellCount; i++)
            {
                if (i * CELL_SIZE + 2 <= data.Length)
                {
                    ushort tileId = BitConverter.ToUInt16(data, i * CELL_SIZE);
                    uniqueTiles.Add(tileId);
                }
            }

            // 为每个唯一 TileId 分配颜色
            var tileColors = new Dictionary<ushort, Color>();
            int colorIndex = 0;
            foreach (var tile in uniqueTiles.OrderBy(t => t))
            {
                tileColors[tile] = GenerateDistinctColor(colorIndex++, uniqueTiles.Count);
            }

            // 绘制单元格
            for (int y = 0; y < gridHeight && y * gridWidth + 0 < cellCount; y++)
            {
                for (int x = 0; x < gridWidth && y * gridWidth + x < cellCount; x++)
                {
                    int idx = y * gridWidth + x;
                    int offset = idx * CELL_SIZE;

                    if (offset + CELL_SIZE <= data.Length)
                    {
                        // 解析 TMapCell
                        ushort tileId = BitConverter.ToUInt16(data, offset);
                        byte tileNum = data[offset + 2];
                        ushort tileOverId = BitConverter.ToUInt16(data, offset + 3);
                        byte boMove = data[offset + 11];

                        // 获取颜色
                        Color cellColor = tileColors.ContainsKey(tileId) ? tileColors[tileId] : Color.Gray;

                        // 根据移动属性调整亮度
                        if (boMove != 0)
                        {
                            cellColor = ControlDarken(cellColor, 0.7f); // 可通行区域稍暗
                        }

                        // 绘制单元格
                        g.FillRectangle(new SolidBrush(cellColor),
                            x * PIXEL_PER_CELL, y * PIXEL_PER_CELL,
                            PIXEL_PER_CELL - 1, PIXEL_PER_CELL - 1);

                        // 绘制叠加层指示（如果有对象）
                        if (tileOverId > 0 || data[offset + 5] > 0)
                        {
                            g.DrawRectangle(Pens.White,
                                x * PIXEL_PER_CELL + 1, y * PIXEL_PER_CELL + 1,
                                PIXEL_PER_CELL - 3, PIXEL_PER_CELL - 3);
                        }
                    }
                }
            }

            picMapPreview.Image = new Bitmap(bmp);
            lblMapInfo.Text = $"网格: {gridWidth}x{gridHeight} | 单元格: {cellCount} | 唯一瓦片: {uniqueTiles.Count}";
        }

        /// <summary>生成区分度高的颜色</summary>
        private static Color GenerateDistinctColor(int index, int total)
        {
            // 使用 HSV 色轮均匀分布
            double hue = (double)index / Math.Max(total, 1) * 360;
            return HsvToRgb(hue, 0.7, 0.9);
        }

        /// <summary>HSV 转 RGB</summary>
        private static Color HsvToRgb(double h, double s, double v)
        {
            double c = v * s;
            double x = c * (1 - Math.Abs((h / 60) % 2 - 1));
            double m = v - c;

            double r = 0, g = 0, b = 0;
            if (h < 60) { r = c; g = x; b = 0; }
            else if (h < 120) { r = x; g = c; b = 0; }
            else if (h < 180) { r = 0; g = c; b = x; }
            else if (h < 240) { r = 0; g = x; b = c; }
            else if (h < 300) { r = x; g = 0; b = c; }
            else { r = c; g = 0; b = x; }

            return Color.FromArgb(
                (int)((r + m) * 255),
                (int)((g + m) * 255),
                (int)((b + m) * 255));
        }

        /// <summary>调暗颜色</summary>
        private static Color ControlDarken(Color color, float factor)
        {
            return Color.FromArgb(
                (int)(color.R * factor),
                (int)(color.G * factor),
                (int)(color.B * factor));
        }
    }

    /// <summary>
    /// DAT 资源包数据结构
    /// </summary>
    internal class DatEntry
    {
        public string Name { get; set; }
        public long Offset { get; set; }
        public int Size { get; set; }
    }

    /// <summary>
    /// DAT 资源包解析器
    /// 支持多种千年游戏 DAT 格式：
    /// - 标准 DAT：固定大小的目录表
    /// - 索引式 DAT：带索引块的格式
    /// - 流式 DAT：连续存储格式
    /// </summary>
    internal class DatPackage
    {
        public List<DatEntry> Entries { get; } = new List<DatEntry>();
        public string PackageType { get; private set; } = "未知";
        public long TotalSize { get; private set; }

        private string _filePath;
        private byte[] _fileData;

        public void Load(string filePath)
        {
            _filePath = filePath;
            _fileData = File.ReadAllBytes(filePath);
            TotalSize = _fileData.Length;

            DetectAndParse();
        }

        private void DetectAndParse()
        {
            // 尝试检测 DAT 格式类型
            if (TryParseStandardDat())
            {
                PackageType = "标准 DAT";
                return;
            }

            if (TryParseIndexedDat())
            {
                PackageType = "索引式 DAT";
                return;
            }

            if (TryParseStreamDat())
            {
                PackageType = "流式 DAT";
                return;
            }

            throw new InvalidDataException("未知的 DAT 格式或损坏的文件");
        }

        private bool TryParseStandardDat()
        {
            // 标准 DAT 格式：
            // [0-3]: 魔数 "DAT\0" 或 "PACK"
            // [4-7]: 版本号或条目数量
            // [8-11]: 目录偏移
            // 之后是数据区

            if (_fileData.Length < 16) return false;

            // 检查常见魔数
            string magic = Encoding.ASCII.GetString(_fileData, 0, Math.Min(4, _fileData.Length));
            if (magic == "DAT\0" || magic == "PACK" || magic == "DAT1" || magic == "DAT2")
            {
                ParseStandardHeader(magic);
                return true;
            }

            // 尝试其他常见格式标识
            if (_fileData[0] == 0x44 && _fileData[1] == 0x41 && _fileData[2] == 0x54) // "DAT"
            {
                ParseStandardHeader(magic);
                return true;
            }

            return false;
        }

        private void ParseStandardHeader(string magic)
        {
            Entries.Clear();

            int offset = 4;

            // 读取版本/参数
            if (magic == "DAT2" || magic == "PACK")
            {
                // 带版本信息的格式
                int version = BitConverter.ToInt32(_fileData, offset); offset += 4;
                int entryCount = BitConverter.ToInt32(_fileData, offset); offset += 4;
                int dirOffset = BitConverter.ToInt32(_fileData, offset); offset += 4;

                ParseDirectory(dirOffset, entryCount);
            }
            else
            {
                // 简单格式
                int entryCount = BitConverter.ToInt32(_fileData, offset); offset += 4;
                int dirOffset = BitConverter.ToInt32(_fileData, offset); offset += 4;

                ParseDirectory(dirOffset, entryCount);
            }
        }

        private void ParseDirectory(int dirOffset, int entryCount)
        {
            if (dirOffset < 0 || dirOffset >= _fileData.Length) return;

            int pos = dirOffset;

            for (int i = 0; i < entryCount && pos + 64 <= _fileData.Length; i++)
            {
                try
                {
                    // 读取文件名（通常为 56 字节）
                    byte[] nameBytes = new byte[56];
                    Array.Copy(_fileData, pos, nameBytes, 0, 56);
                    pos += 56;

                    string name = Encoding.ASCII.GetString(nameBytes).TrimEnd('\0');
                    if (string.IsNullOrEmpty(name)) continue;

                    // 读取偏移和大小
                    int fileOffset = BitConverter.ToInt32(_fileData, pos); pos += 4;
                    int fileSize = BitConverter.ToInt32(_fileData, pos); pos += 4;

                    if (fileOffset > 0 && fileSize > 0 && fileOffset + fileSize <= _fileData.Length)
                    {
                        Entries.Add(new DatEntry
                        {
                            Name = name,
                            Offset = fileOffset,
                            Size = fileSize
                        });
                    }
                }
                catch
                {
                    break;
                }
            }
        }

        private bool TryParseIndexedDat()
        {
            // 索引式 DAT 格式：
            // 文件末尾有索引块，记录每个文件的名称、偏移、大小

            if (_fileData.Length < 256) return false;

            // 在文件末尾查找索引块
            // 通常索引块以特定标记开始或在最后几个KB

            int searchStart = Math.Max(0, _fileData.Length - 65536);

            for (int i = searchStart; i < _fileData.Length - 64; i++)
            {
                // 查找可能的文件名模式（ASCII字符序列）
                if (IsValidFileNameStart(i))
                {
                    if (TryReadIndexBlock(i))
                    {
                        return true;
                    }
                }
            }

            return false;
        }

        private bool IsValidFileNameStart(int offset)
        {
            // 检查是否像文件名开头（字母或数字）
            if (offset + 4 >= _fileData.Length) return false;

            char c = (char)_fileData[offset];
            return char.IsLetterOrDigit(c) || c == '_' || c == '.';
        }

        private bool TryReadIndexBlock(int startOffset)
        {
            // 尝试从这个位置读取索引条目
            List<DatEntry> tempEntries = new List<DatEntry>();
            int pos = startOffset;
            int count = 0;

            while (pos + 72 <= _fileData.Length && count < 1000) // 限制最大条目数
            {
                try
                {
                    // 读取文件名
                    byte[] nameBytes = new byte[64];
                    Array.Copy(_fileData, pos, nameBytes, 0, 64);
                    pos += 64;

                    string name = Encoding.ASCII.GetString(nameBytes).TrimEnd('\0', ' ');
                    if (string.IsNullOrEmpty(name) || !IsValidFileName(name)) break;

                    // 读取偏移和大小
                    int fileOffset = BitConverter.ToInt32(_fileData, pos); pos += 4;
                    int fileSize = BitConverter.ToInt32(_fileData, pos); pos += 4;

                    if (fileOffset > 0 && fileSize > 0 && fileOffset + fileSize <= _fileData.Length &&
                        fileOffset < startOffset) // 数据应该在索引之前
                    {
                        tempEntries.Add(new DatEntry
                        {
                            Name = name,
                            Offset = fileOffset,
                            Size = fileSize
                        });
                        count++;
                    }
                    else
                    {
                        break;
                    }
                }
                catch
                {
                    break;
                }
            }

            // 如果找到了足够多的有效条目，认为这是一个有效的索引块
            if (count >= 5)
            {
                Entries.Clear();
                Entries.AddRange(tempEntries);
                return true;
            }

            return false;
        }

        private bool IsValidFileName(string name)
        {
            // 检查是否看起来像一个有效的文件名
            if (name.Length < 2 || name.Length > 60) return false;

            foreach (char c in name)
            {
                if (!char.IsLetterOrDigit(c) && c != '_' && c != '-' && c != '.' && c != '\\' && c != '/')
                {
                    return false;
                }
            }

            // 应该包含扩展名
            return name.Contains(".");
        }

        private bool TryParseStreamDat()
        {
            // 流式 DAT 格式：
            // 连续存储多个文件，每个文件前有其长度前缀
            // 这种格式没有明确的目录，需要通过启发式方法识别

            Entries.Clear();
            int pos = 0;
            int fileIndex = 0;

            while (pos + 8 <= _fileData.Length && fileIndex < 1000)
            {
                try
                {
                    // 读取可能的长度字段
                    int possibleSize = BitConverter.ToInt32(_fileData, pos);

                    // 合理性检查
                    if (possibleSize > 0 && possibleSize < 10 * 1024 * 1024 && // 单文件不超过10MB
                        pos + 4 + possibleSize <= _fileData.Length)
                    {
                        // 检查下一个位置是否也是合理的开始
                        int nextPos = pos + 4 + possibleSize;
                        if (nextPos + 4 <= _fileData.Length)
                        {
                            int nextSize = BitConverter.ToInt32(_fileData, nextPos);
                            if (nextSize > 0 && nextSize < 10 * 1024 * 1024)
                            {
                                // 可能是一个有效的流式条目
                                Entries.Add(new DatEntry
                                {
                                    Name = $"file_{fileIndex:D4}.bin",
                                    Offset = pos + 4,
                                    Size = possibleSize
                                });

                                pos = nextPos;
                                fileIndex++;
                                continue;
                            }
                        }
                    }
                }
                catch
                {
                    break;
                }

                pos++;
            }

            return Entries.Count >= 3; // 至少找到3个文件才认为是有效的
        }

        public byte[] Extract(string fileName)
        {
            var entry = Entries.FirstOrDefault(e => e.Name.Equals(fileName, StringComparison.OrdinalIgnoreCase));
            if (entry == null)
                throw new FileNotFoundException($"找不到文件: {fileName}");

            if (entry.Offset + entry.Size > _fileData.Length)
                throw new InvalidDataException("文件数据超出范围");

            byte[] data = new byte[entry.Size];
            Array.Copy(_fileData, entry.Offset, data, 0, entry.Size);
            return data;
        }

        public Stream ExtractToStream(string fileName)
        {
            byte[] data = Extract(fileName);
            return new MemoryStream(data);
        }
    }
}
