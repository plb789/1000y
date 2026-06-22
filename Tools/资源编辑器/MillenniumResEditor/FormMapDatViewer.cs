using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Text;
using System.Windows.Forms;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor
{
    /// <summary>
    /// MAP .DAT 地图包查看器
    /// 用于查看和解包千年游戏地图资源包中的地图块数据
    /// </summary>
    public partial class FormMapDatViewer : Form
    {
        private DatPackage _package;
        private MapBlockParser _currentBlock;
        private string _mapFilePath;
        private List<MapDatEntry> _mapEntries = new List<MapDatEntry>();

        // 完整地图数据（从 .dm 文件加载）
        private MapCell[,] _fullMapCells;
        private int _fullMapWidth;
        private int _fullMapHeight;

        // 瓦片图集（用于渲染真实图像）
        private TileLibraryParser _tileLib;
        private string _tileLibPath;

        public FormMapDatViewer()
        {
            InitializeComponent();
        }

        // ==================== 事件处理 ====================

        private void btnOpenMapDat_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog
            {
                Filter = "MAP DAT包(*.dat)|*.dat|所有文件(*.*)|*.*",
                Title = "选择 MAP .dat 地图包"
            };

            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                _package = new DatPackage();
                _package.Load(dlg.FileName);

                _mapEntries.Clear();
                lstBlocks.Items.Clear();

                foreach (var entry in _package.Entries)
                {
                    // 筛选地图相关文件（.DT, .dt, .do 等）
                    if (IsMapFile(entry.Name))
                    {
                        _mapEntries.Add(new MapDatEntry { Entry = entry });
                        lstBlocks.Items.Add($"{entry.Name} ({FormatSize(entry.Size)})");
                    }
                }

                lblInfo.Text = $"包: {Path.GetFileName(dlg.FileName)} | 地图块: {_mapEntries.Count} 个 | 总条目: {_package.Entries.Count}";
                lblMapSize.Text = "地图尺寸: - x - 格";
                lblBlockSize.Text = "块大小: 40 x 40";
                txtDetails.Text = "请选择一个地图块查看详情";
                picPreview.Image?.Dispose();
                picPreview.Image = null;
                lblCurrentBlock.Text = "当前块: -";
            }
            catch (Exception ex)
            {
                MessageBox.Show($"加载失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void btnOpenMap_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog
            {
                Filter = "千年地图(*.dm;*.map)|*.dm;*.map|DM地图文件(*.dm)|*.dm|MAP头文件(*.map)|*.map|所有文件(*.*)|*.*",
                Title = "选择 .dm/.map 地图文件"
            };

            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                _mapFilePath = dlg.FileName;
                string ext = Path.GetExtension(dlg.FileName).ToLower();

                if (ext == ".dm")
                {
                    LoadDmFile(dlg.FileName);
                }
                else
                {
                    var headerParser = new MapBlockParser();
                    headerParser.LoadMapHeader(_mapFilePath);

                    lblInfo.Text = $"地图: {Path.GetFileName(dlg.FileName)} | 标识: {headerParser.MapIdent}";
                    lblMapSize.Text = $"地图尺寸: {headerParser.MapWidth} x {headerParser.MapHeight} 格";
                    lblBlockSize.Text = $"块大小: {headerParser.BlockSize} x {headerParser.BlockSize}";

                    txtDetails.Text = $"=== 地图信息 ===\n" +
                                      $"标识: {headerParser.MapIdent}\n" +
                                      $"宽度: {headerParser.MapWidth} 格\n" +
                                      $"高度: {headerParser.MapHeight} 格\n\n" +
                                      "提示: 请打开对应的 .dm 文件查看完整地图数据";
                }
            }
            catch (Exception ex)
            {
                MessageBox.Show($"加载失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        /// <summary>加载完整 .dm 地图文件</summary>
        private void LoadDmFile(string filePath)
        {
            byte[] data = File.ReadAllBytes(filePath);
            using var ms = new MemoryStream(data);
            using var br = new BinaryReader(ms);

            // .dm 文件结构（来自 CLMap.pas 源码）:
            // [0-15]   : 加密文件名(16B) ← 需要跳过
            // [16-47]  : TMapFileInfo(32B)
            // [48+]    : TMapBlockData[]...

            // 跳过前16字节的加密文件名
            br.ReadBytes(16);

            // 读取 TMapFileInfo（32字节）
            byte[] identBytes = br.ReadBytes(16);
            string mapIdent = Encoding.ASCII.GetString(identBytes).TrimEnd('\0');

            int blockSize = br.ReadInt32();
            int mapWidth = br.ReadInt32();
            int mapHeight = br.ReadInt32();

            // 边界检查
            if (mapWidth <= 0 || mapHeight <= 0 || mapWidth > 10000 || mapHeight > 10000)
            {
                throw new InvalidDataException($"无效的地图尺寸: {mapWidth} x {mapHeight}，标识: {mapIdent}");
            }

            if (blockSize <= 0 || blockSize > 256)
            {
                throw new InvalidDataException($"无效的块大小: {blockSize}，标识: {mapIdent}");
            }

            _fullMapWidth = mapWidth;
            _fullMapHeight = mapHeight;

            int blocksX = (int)Math.Ceiling((double)mapWidth / MapBlockParser.MAP_BLOCK_SIZE);
            int blocksY = (int)Math.Ceiling((double)mapHeight / MapBlockParser.MAP_BLOCK_SIZE);

            _fullMapCells = new MapCell[mapWidth, mapHeight];
            for (int y = 0; y < mapHeight; y++)
                for (int x = 0; x < mapWidth; x++)
                    _fullMapCells[x, y] = new MapCell();

            lstBlocks.Items.Clear();
            _mapEntries.Clear();

            for (int by = 0; by < blocksY; by++)
            {
                for (int bx = 0; bx < blocksX; bx++)
                {
                    try
                    {
                        byte[] blockIdent = br.ReadBytes(16);
                        int changedCount = br.ReadInt32();

                        int startX = bx * MapBlockParser.MAP_BLOCK_SIZE;
                        int startY = by * MapBlockParser.MAP_BLOCK_SIZE;

                        for (int cy = 0; cy < MapBlockParser.MAP_BLOCK_SIZE && startY + cy < mapHeight; cy++)
                        {
                            for (int cx = 0; cx < MapBlockParser.MAP_BLOCK_SIZE && startX + cx < mapWidth; cx++)
                            {
                                _fullMapCells[startX + cx, startY + cy] = new MapCell
                                {
                                    TileId = ReadUInt16(br),
                                    TileNumber = br.ReadByte(),
                                    TileOverId = ReadUInt16(br),
                                    TileOverNumber = br.ReadByte(),
                                    ObjectId = ReadUInt16(br),
                                    ObjectNumber = br.ReadByte(),
                                    RoofId = ReadUInt16(br),
                                    BoMove = br.ReadByte()
                                };
                            }
                        }

                        lstBlocks.Items.Add($"块 ({bx},{by}) | 修改:{changedCount}");
                        _mapEntries.Add(new MapDatEntry { Entry = new DatEntry { Name = $"block_{bx}_{by}" }, BlockX = bx, BlockY = by });
                    }
                    catch (EndOfStreamException)
                    {
                        break;
                    }
                }
            }

            lblInfo.Text = $"地图: {Path.GetFileName(filePath)} | 标识: {mapIdent} | 块数: {_mapEntries.Count}";
            lblMapSize.Text = $"地图尺寸: {mapWidth} x {mapHeight} 格";
            lblBlockSize.Text = $"块大小: {blockSize} x {blockSize}";

            txtDetails.Text = $"=== 完整地图信息 ===\n" +
                              $"文件: {Path.GetFileName(filePath)}\n" +
                              $"标识: {mapIdent}\n" +
                              $"宽度: {mapWidth} 格\n" +
                              $"高度: {mapHeight} 格\n" +
                              $"像素尺寸: ~{mapWidth * 32} x {mapHeight * 24}\n" +
                              $"块数量: {blocksX} x {blocksY} = {blocksX * blocksY}\n\n" +
                              "提示:\n- 点击「导出地图PNG」可导出完整地图图像\n- 选择单个块可预览详情";

            GenerateFullMapPreview(8, 6);
        }

        private static ushort ReadUInt16(BinaryReader br)
        {
            byte[] buf = br.ReadBytes(2);
            return (ushort)(buf[0] | (buf[1] << 8));
        }

        private void btnExtractAll_Click(object sender, EventArgs e)
        {
            if (_package == null || _mapEntries.Count == 0)
            {
                MessageBox.Show("请先打开 MAP .dat 文件", "提示");
                return;
            }

            using var dlg = new FolderBrowserDialog { Description = "选择解包输出目录" };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            string outputDir = Path.Combine(dlg.SelectedPath, "map_blocks");
            Directory.CreateDirectory(outputDir);

            int extracted = 0;
            int failed = 0;

            foreach (var mapEntry in _mapEntries)
            {
                try
                {
                    byte[] data = _package.Extract(mapEntry.Entry.Name);
                    string filePath = Path.Combine(outputDir, mapEntry.Entry.Name);
                    string dir = Path.GetDirectoryName(filePath);
                    if (!Directory.Exists(dir))
                    {
                        Directory.CreateDirectory(dir);
                    }

                    File.WriteAllBytes(filePath, data);
                    extracted++;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[警告] 解包失败: {mapEntry.Entry.Name} - {ex.Message}");
                    failed++;
                }
            }

            MessageBox.Show($"解包完成!\n成功: {extracted}\n失败: {failed}\n\n输出目录: {outputDir}", "解包结果");
        }

        private void btnExportTmx_Click(object sender, EventArgs e)
        {
            if (_currentBlock == null)
            {
                MessageBox.Show("请先选择并预览一个地图块", "提示");
                return;
            }

            using var dlg = new SaveFileDialog
            {
                Filter = "TMX地图(*.tmx)|*.tmx",
                FileName = $"block_{_currentBlock.BlockX}_{_currentBlock.BlockY}.tmx",
                Title = "保存 TMX 文件"
            };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                string tmxContent = _currentBlock.ExportToTmx();
                File.WriteAllText(dlg.FileName, tmxContent, Encoding.UTF8);
                MessageBox.Show($"TMX 已保存: {dlg.FileName}", "导出成功");
            }
            catch (Exception ex)
            {
                MessageBox.Show($"TMX 导出失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void btnPreviewBlock_Click(object sender, EventArgs e)
        {
            if (lstBlocks.SelectedIndex < 0)
            {
                MessageBox.Show("请先选择一个地图块", "提示");
                return;
            }

            PreviewSelectedBlock();
        }

        private void lstBlocks_SelectedIndexChanged(object sender, EventArgs e)
        {
            if (lstBlocks.SelectedIndex < 0) return;

            // 支持 DAT 包 或 完整 .dm 地图
            if (_package == null && _fullMapCells == null) return;

            PreviewSelectedBlock();
        }

        // ==================== 辅助方法 ====================

        private void PreviewSelectedBlock()
        {
            if (lstBlocks.SelectedIndex < 0) return;
            if (_package == null && _fullMapCells == null) return;

            try
            {
                var mapEntry = _mapEntries[lstBlocks.SelectedIndex];

                if (_fullMapCells != null)
                {
                    // 从完整 .dm 地图数组中提取块数据
                    int bx = mapEntry.BlockX;
                    int by = mapEntry.BlockY;

                    var blockParser = new MapBlockParser();
                    blockParser.LoadFromMapArray(_fullMapCells, bx, by);
                    _currentBlock = blockParser;

                    // 生成预览图
                    picPreview.Image?.Dispose();
                    picPreview.Image = _currentBlock.ExportToImage(8, 6);

                    lblCurrentBlock.Text = $"当前块: ({bx}, {by})";
                    lblBlockSize.Text = $"块大小: {MapBlockParser.MAP_BLOCK_SIZE} x {MapBlockParser.MAP_BLOCK_SIZE}";

                    // 显示详情
                    ShowBlockDetails(bx, by);
                }
                else
                {
                    // 从 DAT 包提取
                    byte[] data = _package.Extract(mapEntry.Entry.Name);

                    using var stream = new MemoryStream(data);
                    _currentBlock = new MapBlockParser();

                    ParseBlockCoordinates(mapEntry.Entry.Name, out int blockX, out int blockY);
                    _currentBlock.LoadBlockFromStream(stream, blockX, blockY);

                    // 更新UI
                    lblCurrentBlock.Text = $"当前块: {mapEntry.Entry.Name} ({blockX}, {blockY})";

                    // 显示预览图像
                    var bmp = _currentBlock.ExportToImage(8, 6);
                    picPreview.Image?.Dispose();
                    picPreview.Image = bmp;

                    // 显示详细信息
                    ShowBlockDetails(mapEntry, data);
                }
            }
            catch (Exception ex)
            {
                txtDetails.Text = $"解析失败:\n{ex.Message}";
                picPreview.Image?.Dispose();
                picPreview.Image = null;
            }
        }

        /// <summary>显示从完整地图提取的块详情</summary>
        private void ShowBlockDetails(int bx, int by)
        {
            var sb = new StringBuilder();
            sb.AppendLine($"=== 块坐标: ({bx}, {by}) ===");
            sb.AppendLine($"大小: {MapBlockParser.MAP_BLOCK_SIZE} x {MapBlockParser.MAP_BLOCK_SIZE}");

            if (_currentBlock == null) return;

            int tileCount = 0, overCount = 0, objCount = 0, roofCount = 0, moveableCount = 0;

            for (int y = 0; y < MapBlockParser.MAP_BLOCK_SIZE; y++)
            {
                for (int x = 0; x < MapBlockParser.MAP_BLOCK_SIZE; x++)
                {
                    var cell = _currentBlock.GetCell(x, y);
                    if (cell != null)
                    {
                        if (cell.TileId > 0) tileCount++;
                        if (cell.TileOverId > 0) overCount++;
                        if (cell.ObjectId > 0) objCount++;
                        if (cell.RoofId > 0) roofCount++;
                        if ((cell.BoMove & 0x01) != 0) moveableCount++;
                    }
                }
            }

            sb.AppendLine($"\n--- 统计 ---");
            sb.AppendLine($"地面瓦片: {tileCount}");
            sb.AppendLine($"叠加层: {overCount}");
            sb.AppendLine($"对象: {objCount}");
            sb.AppendLine($"屋顶: {roofCount}");
            sb.AppendLine($"可行走格: {moveableCount}");

            sb.AppendLine("\n=== 前10个单元格 ===");
            for (int i = 0; i < Math.Min(10, MapBlockParser.MAP_BLOCK_SIZE * MapBlockParser.MAP_BLOCK_SIZE); i++)
            {
                int cx = i % MapBlockParser.MAP_BLOCK_SIZE;
                int cy = i / MapBlockParser.MAP_BLOCK_SIZE;
                var cell = _currentBlock.GetCell(cx, cy);
                if (cell != null)
                {
                    sb.AppendLine($"({cx},{cy}): T={cell.TileId}/{cell.TileNumber} " +
                                  $"Ov={cell.TileOverId} Obj={cell.ObjectId} Rf={cell.RoofId} M={cell.BoMove}");
                }
            }

            txtDetails.Text = sb.ToString();
        }

        private void ShowBlockDetails(MapDatEntry mapEntry, byte[] data)
        {
            var sb = new StringBuilder();

            sb.AppendLine("=== 文件信息 ===");
            sb.AppendLine($"名称: {mapEntry.Entry.Name}");
            sb.AppendLine($"大小: {data.Length} 字节");
            sb.AppendLine($"偏移: 0x{mapEntry.Entry.Offset:X8}");

            if (_currentBlock != null)
            {
                sb.AppendLine("\n=== 块信息 ===");
                sb.AppendLine($"坐标: ({_currentBlock.BlockX}, {_currentBlock.BlockY})");
                sb.AppendLine($"尺寸: {_currentBlock.BlockSize} x {_currentBlock.BlockSize}");
                sb.AppendLine($"修改计数: {_currentBlock.ChangedCount}");
                sb.AppendLine("标识: (见文件头)");

                sb.AppendLine("\n=== 单元格统计 ===");

                int tileCount = 0, overCount = 0, objCount = 0, roofCount = 0, moveableCount = 0;

                for (int y = 0; y < _currentBlock.BlockSize; y++)
                {
                    for (int x = 0; x < _currentBlock.BlockSize; x++)
                    {
                        var cell = _currentBlock.GetCell(x, y);
                        if (cell != null)
                        {
                            if (cell.TileId > 0) tileCount++;
                            if (cell.TileOverId > 0) overCount++;
                            if (cell.ObjectId > 0) objCount++;
                            if (cell.RoofId > 0) roofCount++;
                            if ((cell.BoMove & 0x01) != 0) moveableCount++;
                        }
                    }
                }

                sb.AppendLine($"地面瓦片: {tileCount}");
                sb.AppendLine($"叠加层: {overCount}");
                sb.AppendLine($"对象: {objCount}");
                sb.AppendLine($"屋顶: {roofCount}");
                sb.AppendLine($"可行走格: {moveableCount}");

                // 显示前几个单元格的详细信息
                sb.AppendLine("\n=== 前10个单元格示例 ===");
                for (int i = 0; i < Math.Min(10, _currentBlock.BlockSize * _currentBlock.BlockSize); i++)
                {
                    int x = i % _currentBlock.BlockSize;
                    int y = i / _currentBlock.BlockSize;
                    var cell = _currentBlock.GetCell(x, y);
                    if (cell != null)
                    {
                        sb.AppendLine($"({x},{y}): {cell}");
                    }
                }
            }

            txtDetails.Text = sb.ToString();
        }

        private bool IsMapFile(string fileName)
        {
            string ext = Path.GetExtension(fileName).ToLower();
            return ext == ".dt" || ext == ".DT" ||
                   ext == ".do" || ext == ".DO" ||
                   ext == ".dm" || ext == ".DM" ||
                   ext == ".map" || ext == ".MAP" ||
                   fileName.Contains("block") || fileName.Contains("BLOCK") ||
                   fileName.Contains("tile") || fileName.Contains("TILE");
        }

        private void ParseBlockCoordinates(string fileName, out int blockX, out int blockY)
        {
            blockX = 0;
            blockY = 0;

            // 尝试从文件名中提取坐标
            // 常见格式: block_X_Y.DT, X_Y.dt, tileXY.do 等

            string name = Path.GetFileNameWithoutExtension(fileName).ToUpper();

            // 尝试匹配 block_X_Y 或 X_Y 格式
            var match = System.Text.RegularExpressions.Regex.Match(name, @"(?:BLOCK)?[_\-]?(\d+)[_\-](\d+)$");
            if (match.Success)
            {
                int.TryParse(match.Groups[1].Value, out blockX);
                int.TryParse(match.Groups[2].Value, out blockY);
            }
            else
            {
                // 如果无法从文件名推断，使用列表索引计算
                if (_mapEntries.Count > 0 && lstBlocks.SelectedIndex >= 0)
                {
                    int index = lstBlocks.SelectedIndex;
                    // 假设按行优先顺序排列
                    blockX = index % 100; // 假设每行最多100个块
                    blockY = index / 100;
                }
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

        /// <summary>生成完整地图缩略图预览</summary>
        private void GenerateFullMapPreview(int tileWidth, int tileHeight)
        {
            if (_fullMapCells == null) return;

            int imgWidth = _fullMapWidth * tileWidth;
            int imgHeight = _fullMapHeight * tileHeight;

            // 限制预览尺寸
            const int MAX_PREVIEW = 400;
            float scale = 1f;
            if (imgWidth > MAX_PREVIEW || imgHeight > MAX_PREVIEW)
            {
                scale = Math.Min((float)MAX_PREVIEW / imgWidth, (float)MAX_PREVIEW / imgHeight);
                imgWidth = (int)(imgWidth * scale);
                imgHeight = (int)(imgHeight * scale);
            }

            var bmp = new Bitmap(imgWidth, imgHeight, PixelFormat.Format32bppArgb);
            using var g = Graphics.FromImage(bmp);
            g.Clear(Color.Black);

            // 收集唯一瓦片ID用于颜色映射
            var uniqueTiles = new HashSet<ushort>();
            for (int y = 0; y < _fullMapHeight; y++)
                for (int x = 0; x < _fullMapWidth; x++)
                    if (_fullMapCells[x, y].TileId > 0)
                        uniqueTiles.Add(_fullMapCells[x, y].TileId);

            // 为每个瓦片分配颜色
            var tileColors = new Dictionary<ushort, Color>();
            int idx = 0;
            foreach (var t in uniqueTiles.ToList().OrderBy(t => t))
                tileColors[t] = HsvToColor(idx++, uniqueTiles.Count);

            // 绘制地图
            int tw = (int)(tileWidth * scale);
            int th = (int)(tileHeight * scale);

            for (int y = 0; y < _fullMapHeight; y++)
            {
                for (int x = 0; x < _fullMapWidth; x++)
                {
                    var cell = _fullMapCells[x, y];
                    Color color = cell.TileId > 0 && tileColors.ContainsKey(cell.TileId)
                        ? tileColors[cell.TileId] : Color.FromArgb(40, 40, 40);

                    // 有叠加层或对象时加边框
                    if (cell.TileOverId > 0 || cell.ObjectId > 0)
                    {
                        using var pen = new Pen(Color.White, 1);
                        g.DrawRectangle(pen, x * tw, y * th, tw - 1, th - 1);
                    }
                    // 可通行区域稍暗
                    else if (cell.BoMove != 0)
                    {
                        color = ControlDarken(color, 0.7f);
                    }

                    using var brush = new SolidBrush(color);
                    g.FillRectangle(brush, x * tw, y * th, tw - 1, th - 1);
                }
            }

            picPreview.Image?.Dispose();
            picPreview.Image = bmp;
            lblCurrentBlock.Text = $"完整地图预览 ({_fullMapWidth}x{_fullMapHeight})";
        }

        /// <summary>加载瓦片图集 (.dt 文件)</summary>
        private void btnLoadTileLib_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog
            {
                Filter = "瓦片图集(*.dt)|*.dt|所有文件(*.*)|*.*",
                Title = "选择瓦片图集 (.dt)",
                InitialDirectory = @"F:\BaiduNetdiskDownload\求生之路\梦千年online\map"
            };

            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                _tileLib?.Dispose();
                _tileLib = new TileLibraryParser();
                _tileLib.Load(dlg.FileName);
                _tileLibPath = dlg.FileName;

                lblInfo.Text = $"地图: {(_mapFilePath ?? "-")} | 瓦片图集: {Path.GetFileName(dlg.FileName)} | " +
                              $"标识: {_tileLib.LibIdent} | 瓦片数: {_tileLib.TileCount}";

                txtDetails.Text = $"=== 瓦片图集信息 ===\n" +
                                  $"文件: {Path.GetFileName(dlg.FileName)}\n" +
                                  $"标识: {_tileLib.LibIdent}\n" +
                                  $"瓦片总数: {_tileLib.TileCount}\n" +
                                  $"头文件: {(_tileLib.HasHeaderFile ? "有 .dh" : "无")}\n\n";

                // 显示前20个瓦片信息
                for (int i = 0; i < Math.Min(20, _tileLib.Tiles.Count); i++)
                {
                    var t = _tileLib.Tiles[i];
                    txtDetails.Text += $"[{i}] ID={t.TileId} {t.TileWidth}x{t.TileHeight} " +
                                       $"Style={t.Style} Block={t.nBlock}\n";
                }
                if (_tileLib.Tiles.Count > 20)
                    txtDetails.Text += $"\n... 共 {_tileLib.Tiles.Count} 个瓦片\n";

                // 如果已有地图数据，自动刷新预览
                if (_fullMapCells != null)
                    RefreshMapPreview();
            }
            catch (Exception ex)
            {
                MessageBox.Show($"加载瓦片图集失败:\n{ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        /// <summary>刷新地图预览（使用真实瓦片）</summary>
        private void RefreshMapPreview()
        {
            // 刷新块列表选中项的预览
            if (lstBlocks.SelectedIndex >= 0)
                PreviewSelectedBlock();
        }

        /// <summary>导出完整地图为 PNG</summary>
        private void btnExportPng_Click(object sender, EventArgs e)
        {
            if (_fullMapCells == null && _currentBlock == null)
            {
                MessageBox.Show("请先打开 .dm 地图文件或选择一个块", "提示");
                return;
            }

            using var dlg = new SaveFileDialog
            {
                Filter = "PNG图像(*.png)|*.png",
                FileName = _mapFilePath != null ? Path.GetFileNameWithoutExtension(_mapFilePath) + "_map" : "map_export",
                Title = "保存地图图像"
            };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                Bitmap bmp;
                if (_fullMapCells != null)
                {
                    // 导出完整地图（高分辨率）
                    bmp = RenderFullMapImage(4, 3); // 每格 4x3 像素（缩小以节省内存）
                }
                else
                {
                    // 导出单个块
                    bmp = _currentBlock.ExportToImage(8, 6);
                }

                bmp.Save(dlg.FileName, ImageFormat.Png);
                bmp.Dispose();

                MessageBox.Show($"地图已导出:\n{dlg.FileName}\n\n尺寸: {_fullMapWidth * 4} x {_fullMapHeight * 3} 像素", "导出成功");
            }
            catch (Exception ex)
            {
                MessageBox.Show($"导出失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        /// <summary>渲染完整地图图像（优先使用真实瓦片）</summary>
        private Bitmap RenderFullMapImage(int tileWidth, int tileHeight)
        {
            int imgWidth = _fullMapWidth * tileWidth;
            int imgHeight = _fullMapHeight * tileHeight;

            var bmp = new Bitmap(imgWidth, imgHeight, PixelFormat.Format32bppArgb);
            using var g = Graphics.FromImage(bmp);
            g.Clear(Color.Black);

            // 如果有瓦片图集，使用真实图像渲染
            if (_tileLib != null && _tileLib.Tiles.Count > 0)
                return RenderMapWithRealTiles(bmp, g, tileWidth, tileHeight);

            // 否则使用纯色模式（原有逻辑）
            return RenderMapWithColorMode(bmp, g, tileWidth, tileHeight);
        }

        /// <summary>使用真实瓦片图集渲染地图</summary>
        private Bitmap RenderMapWithRealTiles(Bitmap bmp, Graphics g, int tileWidth, int tileHeight)
        {
            var tileCache = new Dictionary<int, Bitmap>();

            try
            {
                for (int y = 0; y < _fullMapHeight; y++)
                {
                    for (int x = 0; x < _fullMapWidth; x++)
                    {
                        var cell = _fullMapCells[x, y];
                        int px = x * tileWidth;
                        int py = y * tileHeight;

                        // 绘制地面层
                        if (cell.TileId > 0)
                        {
                            if (!tileCache.TryGetValue(cell.TileId, out var tileBmp))
                            {
                                tileBmp = _tileLib.ReadTileImageDirect(cell.TileId, cell.TileNumber);
                                tileCache[cell.TileId] = tileBmp;
                            }

                            if (tileBmp != null)
                            {
                                g.DrawImage(tileBmp, px, py, tileWidth, tileHeight);
                            }
                            else
                            {
                                // 瓦片不存在，用颜色标记
                                using var brush = new SolidBrush(Color.FromArgb(80, 40, 40));
                                g.FillRectangle(brush, px, py, tileWidth, tileHeight);
                                g.DrawString($"{cell.TileId}", SystemFonts.DefaultFont, Brushes.White, px + 1, py + 1);
                            }
                        }

                        // 绘制叠加层
                        if (cell.TileOverId > 0)
                        {
                            if (!tileCache.TryGetValue(cell.TileOverId, out var overBmp))
                            {
                                overBmp = _tileLib.ReadTileImageDirect(cell.TileOverId, cell.TileOverNumber);
                                tileCache[cell.TileOverId] = overBmp;
                            }
                            if (overBmp != null)
                                g.DrawImage(overBmp, px, py, tileWidth, tileHeight);
                        }
                    }
                }
            }
            finally
            {
                // 释放缓存
                foreach (var t in tileCache.Values) t?.Dispose();
            }

            return bmp;
        }

        /// <summary>使用纯色模式渲染地图（无瓦片图集时的后备方案）</summary>
        private Bitmap RenderMapWithColorMode(Bitmap bmp, Graphics g, int tileWidth, int tileHeight)
        {
            // 收集唯一瓦片
            var uniqueTiles = new HashSet<ushort>();
            for (int y = 0; y < _fullMapHeight; y++)
                for (int x = 0; x < _fullMapWidth; x++)
                    if (_fullMapCells[x, y].TileId > 0)
                        uniqueTiles.Add(_fullMapCells[x, y].TileId);

            var tileColors = new Dictionary<ushort, Color>();
            int colorIdx = 0;
            foreach (var t in uniqueTiles.ToList().OrderBy(t => t))
                tileColors[t] = HsvToColor(colorIdx++, uniqueTiles.Count);

            g.Clear(Color.Black);

            for (int y = 0; y < _fullMapHeight; y++)
            {
                for (int x = 0; x < _fullMapWidth; x++)
                {
                    var cell = _fullMapCells[x, y];
                    Color color = cell.TileId > 0 && tileColors.ContainsKey(cell.TileId)
                        ? tileColors[cell.TileId] : Color.FromArgb(30, 30, 30);

                    if (cell.BoMove != 0)
                        color = ControlDarken(color, 0.75f);

                    using var brush = new SolidBrush(color);
                    g.FillRectangle(brush, x * tileWidth, y * tileHeight, tileWidth, tileHeight);

                    // 叠加层标记
                    if (cell.TileOverId > 0 || cell.ObjectId > 0)
                    {
                        using var pen = new Pen(Color.FromArgb(180, 255, 255, 255), 1);
                        g.DrawRectangle(pen, x * tileWidth + 1, y * tileHeight + 1,
                            tileWidth - 2, tileHeight - 2);
                    }
                }
            }

            return bmp;
        }

        private static Color HsvToColor(int index, int total)
        {
            double hue = total > 1 ? (double)index / (total - 1) * 360 : 0;
            double c = 0.7 * 0.9;
            double x = c * (1 - Math.Abs((hue / 60) % 2 - 1));
            double m = 0.9 - c;

            double r = 0, g = 0, b = 0;
            if (hue < 60) { r = c; g = x; }
            else if (hue < 120) { r = x; g = c; }
            else if (hue < 180) { g = c; b = x; }
            else if (hue < 240) { g = x; b = c; }
            else if (hue < 300) { r = x; b = c; }
            else { r = c; b = x; }

            return Color.FromArgb(
                (int)((r + m) * 255),
                (int)((g + m) * 255),
                (int)((b + m) * 255));
        }

        private static Color ControlDarken(Color color, float factor)
        {
            return Color.FromArgb(
                (int)(color.R * factor),
                (int)(color.G * factor),
                (int)(color.B * factor));
        }
    }

    /// <summary>
    /// MAP DAT 条目包装类
    /// </summary>
    internal class MapDatEntry
    {
        public DatEntry Entry { get; set; }
        public int BlockX { get; set; }
        public int BlockY { get; set; }
    }
}
