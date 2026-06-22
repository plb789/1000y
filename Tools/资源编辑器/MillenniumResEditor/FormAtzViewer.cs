using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Windows.Forms;
using MillenniumResEditor.Converter;
using MillenniumResEditor.Model;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor
{
    /// <summary>
    /// .atz 文件预览与导出窗体
    /// 支持三种模式：
    ///   1. 单个 .atz 文件预览（从独立文件）
    ///   2. 从 PGK 包提取并预览
    ///   3. 批量合成导出 PNG 序列帧
    /// </summary>
    public partial class FormAtzViewer : Form
    {
        private AtzParser _atzParser;
        private int _currentIndex = 0;
        private PgkExtractor _pgkExtractor;
        private Timer tmrAnimation;
        private bool isPlaying = false;

        public FormAtzViewer()
        {
            InitializeComponent();

            // 初始化动画定时器
            tmrAnimation = new Timer();
            tmrAnimation.Tick += TmrAnimation_Tick;
        }

        public FormAtzViewer(string filePath) : this()
        {
            LoadAtzFile(filePath);
        }

        /// <summary>打开单个 .atz 文件</summary>
        private void btnOpenAtz_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog
            {
                Filter = "ATZ 文件(*.atz)|*.atz|DAT 文件(*.dat)|*.dat|所有文件(*.*)|*.*",
                Title = "选择 .atz 文件"
            };
            if (dlg.ShowDialog() == DialogResult.OK)
            {
                LoadAtzFile(dlg.FileName);
            }
        }

        /// <summary>从 PGK 包提取</summary>
        private void btnOpenPgk_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog
            {
                Filter = "PGK/DGK 包(*.pgk;*.dgk)|*.pgk;*.dgk|所有文件(*.*)|*.*",
                Title = "选择 sprite.pgk 或 sys.pgk"
            };
            if (dlg.ShowDialog() == DialogResult.OK)
            {
                try
                {
                    _pgkExtractor = new PgkExtractor();
                    _pgkExtractor.Load(dlg.FileName);

                    lstPgkFiles.Items.Clear();
                    foreach (var entry in _pgkExtractor.Files)
                    {
                        lstPgkFiles.Items.Add(entry.RealName);
                    }

                    lblInfo.Text = $"已加载 {dlg.FileName}，共 {_pgkExtractor.Files.Count} 个文件";
                }
                catch (Exception ex)
                {
                    MessageBox.Show($"加载 PGK 包失败：{ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
                }
            }
        }

        /// <summary>从 PGK 包中选择文件后加载</summary>
        private void lstPgkFiles_SelectedIndexChanged(object sender, EventArgs e)
        {
            if (lstPgkFiles.SelectedIndex < 0 || _pgkExtractor == null) return;

            string realName = lstPgkFiles.SelectedItem.ToString();
            string ext = Path.GetExtension(realName).ToUpperInvariant();

            try
            {
                using var ms = _pgkExtractor.Extract(realName);
                if (ms == null)
                {
                    MessageBox.Show($"无法提取文件：{realName}", "提示");
                    return;
                }

                if (ext == ".ATZ")
                {
                    // .ATZ 动画文件：用 AtzParser 解析
                    var (diagMs, diagInfo) = _pgkExtractor.ExtractWithDiag(realName);
                    if (diagMs == null)
                    {
                        MessageBox.Show($"无法提取文件：{realName}\n{diagInfo}", "提示");
                        return;
                    }

                    string logPath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "atz_diag.log");
                    File.WriteAllText(logPath, $"=== {realName} ===\n{diagInfo}\n");

                    _atzParser = new AtzParser();
                    try
                    {
                        _atzParser.LoadFromStream(diagMs);
                        ShowAtzInfo($"{realName}");
                        File.AppendAllText(logPath, "解析成功!\n");
                    }
                    catch (Exception ex)
                    {
                        File.AppendAllText(logPath, $"异常: {ex.Message}\n{ex.StackTrace}\n");
                        MessageBox.Show(
                            $"{ex.Message}\n\n详细日志: {logPath}",
                            "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
                    }
                }
                else if (ext == ".ATD")
                {
                    // .ATD 动画帧序列定义：用 AtdParser 解析（文本数据库格式）
                    using var atdStream = _pgkExtractor.Extract(realName);
                    if (atdStream == null)
                    {
                        MessageBox.Show($"无法提取文件：{realName}", "提示");
                        return;
                    }

                    try
                    {
                        var atdParser = new AtdParser();
                        byte[] atdData = new byte[atdStream.Length];
                        atdStream.Read(atdData, 0, atdData.Length);

                        // ATD 是文本数据库，先尝试不解密，再尝试解密
                        bool parsed = false;

                        // 方式1：直接解析（PGK内可能已解密）
                        try
                        {
                            atdParser.LoadFromStream(new MemoryStream(atdData));
                            parsed = true;
                        }
                        catch
                        {
                            // 方式2：ROL5 解密后解析
                            PgkExtractor.Encryption(atdData, atdData.Length);
                            try
                            {
                                atdParser.LoadFromStream(new MemoryStream(atdData));
                                parsed = true;
                            }
                            catch { }
                        }

                        if (parsed)
                        {
                            _atzParser = null; // ATD 不用 AtzParser
                            lblInfo.Text = $"{realName} | ATD 动画序列 | {atdParser.Animations.Count} 个动作";
                            lstFrames.Items.Clear();
                            foreach (var anim in atdParser.Animations)
                                lstFrames.Items.Add($"{anim.Action} {anim.Direction} 帧:{anim.Frames?.Count ?? 0}");
                            trkFrame.Enabled = false;
                            picPreview.Image?.Dispose();
                            picPreview.Image = null;
                        }
                        else
                        {
                            lblInfo.Text = $"{realName} | ATD 解析失败（可能需要其他解密方式）";
                        }
                    }
                    catch (Exception ex)
                    {
                        MessageBox.Show($"ATD 解析失败：{ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
                    }
                }
                else if (ext == ".BMP")
                {
                    // .BMP 位图文件：直接显示
                    _atzParser = null;
                    lblInfo.Text = $"{realName} | 格式：BMP 位图";
                    lstFrames.Items.Clear();
                    trkFrame.Enabled = false;

                    using var bmp = new System.Drawing.Bitmap(ms);
                    picPreview.Image?.Dispose();
                    picPreview.Image = bmp.Clone() as System.Drawing.Bitmap;
                }
                else
                {
                    // 其他格式暂不支持预览
                    _atzParser = null;
                    lblInfo.Text = $"{realName} | 格式：{ext}（暂不支持预览）";
                    lstFrames.Items.Clear();
                    trkFrame.Enabled = false;
                    picPreview.Image?.Dispose();
                    picPreview.Image = null;
                }
            }
            catch (Exception ex)
            {
                MessageBox.Show($"解析失败：{ex.Message}\n\n{ex.StackTrace}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        /// <summary>加载 .atz 文件</summary>
        private void LoadAtzFile(string filePath)
        {
            try
            {
                _atzParser = new AtzParser();
                _atzParser.Load(filePath);
                ShowAtzInfo(Path.GetFileName(filePath));
            }
            catch (Exception ex)
            {
                MessageBox.Show($"加载失败：{ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        /// <summary>显示 .atz 信息</summary>
        private void ShowAtzInfo(string fileName)
        {
            lblInfo.Text = $"{fileName} | 格式：{_atzParser.FormatId} | 帧数：{_atzParser.Frames.Count}";

            lstFrames.Items.Clear();
            for (int i = 0; i < _atzParser.Frames.Count; i++)
            {
                var f = _atzParser.Frames[i];
                lstFrames.Items.Add($"#{i} {f.Width}x{f.Height} ({f.OffsetX},{f.OffsetY})");
            }

            if (_atzParser.Frames.Count > 0)
            {
                trkFrame.Maximum = _atzParser.Frames.Count - 1;
                trkFrame.Enabled = true;
                trkFrame.Value = 0;
                ShowFrame(0);
            }
        }

        private void lstFrames_SelectedIndexChanged(object sender, EventArgs e)
        {
            if (lstFrames.SelectedIndex >= 0)
            {
                ShowFrame(lstFrames.SelectedIndex);
                trkFrame.Value = lstFrames.SelectedIndex;
            }
        }

        private void trkFrame_Scroll(object sender, EventArgs e)
        {
            ShowFrame(trkFrame.Value);
            if (trkFrame.Value < lstFrames.Items.Count)
                lstFrames.SelectedIndex = trkFrame.Value;
        }

        /// <summary>显示指定帧</summary>
        private void ShowFrame(int index)
        {
            if (_atzParser == null || index < 0 || index >= _atzParser.Frames.Count) return;

            _currentIndex = index;
            var frame = _atzParser.Frames[index];
            using var bmp = AtzParser.FrameToBitmap(frame);

            // 棋盘格背景便于查看透明
            var displayBmp = new Bitmap(bmp.Width, bmp.Height);
            using (var g = Graphics.FromImage(displayBmp))
            {
                // 绘制棋盘格
                int gridSize = 16;
                for (int y = 0; y < displayBmp.Height; y += gridSize)
                {
                    for (int x = 0; x < displayBmp.Width; x += gridSize)
                    {
                        Color c = ((x / gridSize + y / gridSize) % 2 == 0) ? Color.LightGray : Color.DarkGray;
                        g.FillRectangle(new SolidBrush(c), x, y, gridSize, gridSize);
                    }
                }
                g.DrawImage(bmp, 0, 0);
            }

            picPreview.Image?.Dispose();
            picPreview.Image = displayBmp;
        }

        /// <summary>导出全部 PNG</summary>
        private void btnExportPng_Click(object sender, EventArgs e)
        {
            if (_atzParser == null || _atzParser.Frames.Count == 0)
            {
                MessageBox.Show("请先加载 .atz 文件", "提示");
                return;
            }

            using var dlg = new FolderBrowserDialog { Description = "选择 PNG 导出目录" };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                string prefix = "frame";
                for (int i = 0; i < _atzParser.Frames.Count; i++)
                {
                    using var bmp = AtzParser.FrameToBitmap(_atzParser.Frames[i]);
                    string path = Path.Combine(dlg.SelectedPath, $"{prefix}_{i:000}.png");
                    bmp.Save(path, ImageFormat.Png);
                }

                MessageBox.Show($"已导出 {_atzParser.Frames.Count} 帧 PNG 到：\n{dlg.SelectedPath}",
                    "导出完成", MessageBoxButtons.OK, MessageBoxIcon.Information);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"导出失败：{ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        /// <summary>批量合成角色动画</summary>
        private void btnExportCharacter_Click(object sender, EventArgs e)
        {
            using var form = new FormCharacterExporter();
            form.ShowDialog();
        }

        // ==================== 新增功能事件处理 ====================

        /// <summary>导出当前帧为 PNG</summary>
        private void btnExportCurrentFrame_Click(object sender, EventArgs e)
        {
            if (_atzParser == null || _currentIndex < 0 || _currentIndex >= _atzParser.Frames.Count)
            {
                MessageBox.Show("请先加载 .atz 文件并选择帧", "提示");
                return;
            }

            using var dlg = new SaveFileDialog
            {
                Filter = "PNG 文件(*.png)|*.png|BMP 文件(*.bmp)|*.bmp",
                FileName = $"frame_{_currentIndex:000}.png",
                Title = "保存当前帧"
            };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                using var bmp = AtzParser.FrameToBitmap(_atzParser.Frames[_currentIndex]);
                ImageFormat fmt = dlg.FilterIndex == 1 ? ImageFormat.Png : ImageFormat.Bmp;
                bmp.Save(dlg.FileName, fmt);
                MessageBox.Show($"已保存：{dlg.FileName}", "导出成功");
            }
            catch (Exception ex)
            {
                MessageBox.Show($"导出失败：{ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        /// <summary>批量导出角色库</summary>
        private void btnBatchExport_Click(object sender, EventArgs e)
        {
            var form = new FormBatchExport();
            form.ShowDialog();
        }

        /// <summary>播放动画</summary>
        private void btnPlayAnimation_Click(object sender, EventArgs e)
        {
            if (_atzParser == null || _atzParser.Frames.Count == 0)
            {
                MessageBox.Show("请先加载 .atz 文件", "提示");
                return;
            }

            isPlaying = true;
            int fps = (int)numFps.Value;
            tmrAnimation.Interval = 1000 / fps;
            tmrAnimation.Start();

            btnPlayAnimation.Enabled = false;
            btnStopAnimation.Enabled = true;
            trkFrame.Enabled = false;
        }

        /// <summary>停止动画</summary>
        private void btnStopAnimation_Click(object sender, EventArgs e)
        {
            StopAnimation();
        }

        private void StopAnimation()
        {
            isPlaying = false;
            tmrAnimation.Stop();
            btnPlayAnimation.Enabled = true;
            btnStopAnimation.Enabled = false;
            trkFrame.Enabled = _atzParser != null && _atzParser.Frames.Count > 0;
        }

        /// <summary>动画定时器 Tick</summary>
        private void TmrAnimation_Tick(object sender, EventArgs e)
        {
            if (_atzParser == null || _atzParser.Frames.Count == 0)
            {
                StopAnimation();
                return;
            }

            _currentIndex++;
            if (_currentIndex >= _atzParser.Frames.Count)
            {
                if (chkLoop.Checked)
                    _currentIndex = 0;
                else
                {
                    _currentIndex = _atzParser.Frames.Count - 1;
                    StopAnimation();
                    return;
                }
            }

            ShowFrame(_currentIndex);
            trkFrame.Value = _currentIndex;
            lstFrames.SelectedIndex = _currentIndex;
        }

        /// <summary>导出为 GIF 动画</summary>
        private void btnExportGif_Click(object sender, EventArgs e)
        {
            if (_atzParser == null || _atzParser.Frames.Count == 0)
            {
                MessageBox.Show("请先加载 .atz 文件", "提示");
                return;
            }

            using var saveDlg = new SaveFileDialog
            {
                Filter = "GIF 动画(*.gif)|*.gif",
                FileName = "animation.gif",
                Title = "保存 GIF 动画"
            };
            if (saveDlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                // 使用简单方式：逐帧编码 GIF（需要 System.Drawing.Common）
                // 由于 .NET 的 GDI+ 不直接支持多帧 GIF 编码，这里使用替代方案：
                // 导出 PNG 序列，用户可用其他工具转换为 GIF

                string dir = Path.GetDirectoryName(saveDlg.FileName);
                string baseName = Path.GetFileNameWithoutExtension(saveDlg.FileName);

                for (int i = 0; i < _atzParser.Frames.Count; i++)
                {
                    using var bmp = AtzParser.FrameToBitmap(_atzParser.Frames[i]);
                    string path = Path.Combine(dir, $"{baseName}_{i:000}.png");
                    bmp.Save(path, ImageFormat.Png);
                }

                MessageBox.Show(
                    $"已导出 {_atzParser.Frames.Count} 帧 PNG 到：\n{dir}\n\n" +
                    "提示：可使用 FFmpeg 合成 GIF:\n" +
                    $"ffmpeg -framerate {(int)numFps.Value} -i \"{baseName}_%%03d.png\" -vf \"split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse\" \"{baseName}.gif\"",
                    "导出完成（PNG序列）",
                    MessageBoxButtons.OK, MessageBoxIcon.Information);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"导出失败：{ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }
    }
}
