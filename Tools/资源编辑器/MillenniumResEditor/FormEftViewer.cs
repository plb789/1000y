using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Windows.Forms;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor
{
    /// <summary>
    /// EFT 特效查看与导出窗体
    /// 用于预览和导出千年游戏特效/技能动画
    /// </summary>
    public partial class FormEftViewer : Form
    {
        private EftParser _eftParser;
        private int _currentIndex = 0;
        private PgkExtractor _pgkExtractor;
        private Timer tmrAnimation;
        private bool isPlaying = false;

        public FormEftViewer()
        {
            InitializeComponent();
            tmrAnimation = new Timer();
            tmrAnimation.Tick += TmrAnimation_Tick;
        }

        // ==================== 事件处理 ====================

        private void btnOpenEft_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog
            {
                Filter = "EFT特效(*.eft)|*.eft|所有文件(*.*)|*.*",
                Title = "选择 EFT 特效文件"
            };

            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                _eftParser = new EftParser();
                _eftParser.Load(dlg.FileName);

                LoadFramesToList();
                lblInfo.Text = $"文件: {Path.GetFileName(dlg.FileName)} | 格式: {_eftParser.FormatId} | 帧数: {_eftParser.Frames.Count}";
            }
            catch (Exception ex)
            {
                MessageBox.Show($"加载失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void btnOpenPgk_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog
            {
                Filter = "PGK包(*.pgk)|*.pgk|所有文件(*.*)|*.*",
                Title = "选择 eft.pgk 包"
            };

            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                _pgkExtractor = new PgkExtractor();
                _pgkExtractor.Load(dlg.FileName);

                lstPgkFiles.Items.Clear();
                foreach (var entry in _pgkExtractor.Files)
                {
                    if (entry.Name.EndsWith(".eft", StringComparison.OrdinalIgnoreCase))
                    {
                        lstPgkFiles.Items.Add(entry.Name);
                    }
                }

                lblInfo.Text = $"包: {Path.GetFileName(dlg.FileName)} | EFT 文件: {lstPgkFiles.Items.Count} 个";
            }
            catch (Exception ex)
            {
                MessageBox.Show($"加载 PGK 包失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void btnExportPng_Click(object sender, EventArgs e)
        {
            if (_eftParser?.Frames == null || _eftParser.Frames.Count == 0)
            {
                MessageBox.Show("请先打开 EFT 文件", "提示");
                return;
            }

            using var dlg = new FolderBrowserDialog { Description = "选择 PNG 导出目录" };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            string dir = Path.Combine(dlg.SelectedPath, "eft_frames");
            Directory.CreateDirectory(dir);

            int exported = 0;
            for (int i = 0; i < _eftParser.Frames.Count; i++)
            {
                try
                {
                    var frame = _eftParser.Frames[i];
                    var bmp = EftParser.FrameToBitmap(frame);
                    string file = Path.Combine(dir, $"frame_{i:D4}.png");
                    bmp.Save(file, ImageFormat.Png);
                    bmp.Dispose();
                    exported++;
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"[警告] 帧 {i} 导出失败: {ex.Message}");
                }
            }

            MessageBox.Show($"已导出 {exported}/{_eftParser.Frames.Count} 帧到:\n{dir}", "导出完成");
        }

        private void btnExportGif_Click(object sender, EventArgs e)
        {
            if (_eftParser?.Frames == null || _eftParser.Frames.Count == 0)
            {
                MessageBox.Show("请先打开 EFT 文件", "提示");
                return;
            }

            using var dlg = new SaveFileDialog
            {
                Filter = "GIF动画(*.gif)|*.gif",
                FileName = "effect.gif",
                Title = "保存 GIF 动画"
            };
            if (dlg.ShowDialog() != DialogResult.OK) return;

            try
            {
                int delay = (int)(1000 / numFps.Value);
                int width = 0, height = 0;

                foreach (var frame in _eftParser.Frames)
                {
                    width = Math.Max(width, frame.Width + Math.Abs(frame.OffsetX));
                    height = Math.Max(height, frame.Height + Math.Abs(frame.OffsetY));
                }

                width = Math.Max(width, 160);
                height = Math.Max(height, 160);

                using var gif = new Bitmap(width * 2, height * 2, PixelFormat.Format8bppIndexed);
                var pal = gif.Palette;

                for (int i = 0; i < 256; i++)
                {
                    pal.Entries[i] = Color.FromArgb(i, i, i);
                }
                pal.Entries[255] = Color.Transparent;
                gif.Palette = pal;

                var frames = new List<Bitmap>();
                foreach (var eftFrame in _eftParser.Frames)
                {
                    var bmp = new Bitmap(gif.Width, gif.Height, PixelFormat.Format32bppArgb);
                    using var g = Graphics.FromImage(bmp);
                    g.Clear(Color.Transparent);

                    var srcBmp = EftParser.FrameToBitmap(eftFrame);
                    g.DrawImage(srcBmp, width / 2 + eftFrame.OffsetX, height / 2 + eftFrame.OffsetY);
                    srcBmp.Dispose();

                    frames.Add(bmp);
                }

                SaveGif(dlg.FileName, frames, delay, 0, 0);
                foreach (var f in frames) f.Dispose();

                MessageBox.Show($"GIF 已保存: {dlg.FileName}\n帧数: {_eftParser.Frames.Count}, FPS: {numFps.Value}", "导出成功");
            }
            catch (Exception ex)
            {
                MessageBox.Show($"GIF 导出失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void lstPgkFiles_SelectedIndexChanged(object sender, EventArgs e)
        {
            if (lstPgkFiles.SelectedIndex < 0 || _pgkExtractor == null) return;

            try
            {
                string name = lstPgkFiles.SelectedItem.ToString();
                using var stream = _pgkExtractor.Extract(name);

                _eftParser = new EftParser();
                _eftParser.LoadFromStream(stream);

                LoadFramesToList();
                lblInfo.Text = $"文件: {name} | 格式: {_eftParser.FormatId} | 帧数: {_eftParser.Frames.Count}";
            }
            catch (Exception ex)
            {
                MessageBox.Show($"解析失败: {ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void lstFrames_SelectedIndexChanged(object sender, EventArgs e)
        {
            if (_eftParser?.Frames == null || lstFrames.SelectedIndex < 0) return;

            _currentIndex = lstFrames.SelectedIndex;
            trkFrame.Value = _currentIndex;

            ShowFrame(_currentIndex);
        }

        private void trkFrame_Scroll(object sender, EventArgs e)
        {
            if (_eftParser?.Frames == null || trkFrame.Value >= _eftParser.Frames.Count) return;

            _currentIndex = trkFrame.Value;
            lstFrames.SelectedIndex = _currentIndex;

            ShowFrame(_currentIndex);
        }

        private void btnPlayAnimation_Click(object sender, EventArgs e)
        {
            if (_eftParser?.Frames == null || _eftParser.Frames.Count <= 1) return;

            tmrAnimation.Interval = (int)(1000 / numFps.Value);
            tmrAnimation.Start();
            isPlaying = true;
            btnPlayAnimation.Enabled = false;
            btnStopAnimation.Enabled = true;
        }

        private void btnStopAnimation_Click(object sender, EventArgs e)
        {
            tmrAnimation.Stop();
            isPlaying = false;
            btnPlayAnimation.Enabled = true;
            btnStopAnimation.Enabled = false;
        }

        private void TmrAnimation_Tick(object sender, EventArgs e)
        {
            if (_eftParser?.Frames == null || _eftParser.Frames.Count == 0) return;

            _currentIndex++;
            if (_currentIndex >= _eftParser.Frames.Count)
            {
                _currentIndex = 0;
            }

            trkFrame.Value = _currentIndex;
            lstFrames.SelectedIndex = _currentIndex;
            ShowFrame(_currentIndex);
        }

        // ==================== 辅助方法 ====================

        private void LoadFramesToList()
        {
            lstFrames.Items.Clear();

            for (int i = 0; i < _eftParser.Frames.Count; i++)
            {
                var frame = _eftParser.Frames[i];
                lstFrames.Items.Add($"[{i:D3}] ID={frame.FrameId} {frame.Width}x{frame.Height} ({frame.OffsetX},{frame.OffsetY})");
            }

            trkFrame.Maximum = Math.Max(0, _eftParser.Frames.Count - 1);
            trkFrame.Enabled = _eftParser.Frames.Count > 1;

            if (_eftParser.Frames.Count > 0)
            {
                _currentIndex = 0;
                lstFrames.SelectedIndex = 0;
                ShowFrame(0);
            }
        }

        private void ShowFrame(int index)
        {
            if (index < 0 || index >= _eftParser.Frames.Count) return;

            var frame = _eftParser.Frames[index];
            var bmp = EftParser.FrameToBitmap(frame);

            picPreview.Image?.Dispose();
            picPreview.Image = bmp;
        }

        private static void SaveGif(string path, List<Bitmap> frames, int delay, int repeat, int transIdx)
        {
            using var fs = new FileStream(path, FileMode.Create);
            using var bw = new BinaryWriter(fs);

            bw.Write((byte)'G');
            bw.Write((byte)'I');
            bw.Write((byte)'F');
            bw.Write((byte)'8');
            bw.Write((byte)'9');
            bw.Write((byte)'a');

            var firstFrame = frames[0];
            bw.Write((short)firstFrame.Width);
            bw.Write((short)firstFrame.Height);
            byte packed = 0x80;
            packed |= 0x07;
            bw.Write(packed);
            bw.Write((byte)0);
            bw.Write((byte)0);

            if (repeat >= 0)
            {
                bw.Write((byte)0x21);
                bw.Write((byte)0xFF);
                bw.Write((byte)11);
                bw.Write(new byte[] { (byte)'N', (byte)'E', (byte)'T', (byte)'S', (byte)'C', (byte)'A', (byte)'P', (byte)'E', (byte)'2', (byte)'.', (byte)'0' });
                bw.Write((byte)3);
                bw.Write((byte)1);
                bw.Write((byte)(repeat & 0xFF));
                bw.Write((byte)((repeat >> 8) & 0xFF));
                bw.Write((byte)0);
            }

            for (int i = 0; i < frames.Count; i++)
            {
                var frame = frames[i];

                bw.Write((byte)0x2C);
                bw.Write((short)0);
                bw.Write((short)0);
                bw.Write((short)frame.Width);
                bw.Write((short)frame.Height);

                byte lctFlag = 0x00;
                if (i == 0)
                {
                    lctFlag = 0x80;
                    lctFlag |= 0x07;
                }

                bw.Write(lctFlag);

                if (i == 0)
                {
                    var pal = frame.Palette;
                    for (int c = 0; c < 256; c++)
                    {
                        if (c < pal.Entries.Length)
                        {
                            bw.Write(pal.Entries[c].R);
                            bw.Write(pal.Entries[c].G);
                            bw.Write(pal.Entries[c].B);
                        }
                        else
                        {
                            bw.Write((byte)0);
                            bw.Write((byte)0);
                            bw.Write((byte)0);
                        }
                    }
                }

                bw.Write((byte)0x08);

                int lzwMinCodeSize = 2;
                byte[] pixelData = GetPixelData(frame);
                byte[] compressed = LzwCompress(pixelData, lzwMinCodeSize + 1);

                bw.Write(compressed);
                bw.Write((byte)0);

                if (frames.Count > 1 && i < frames.Count - 1)
                {
                    bw.Write((byte)0x21);
                    bw.Write((byte)0xF9);
                    bw.Write((byte)4);
                    bw.Write((byte)0x04);
                    bw.Write((byte)(delay & 0xFF));
                    bw.Write((byte)((delay >> 8) & 0xFF));
                    bw.Write(transIdx > 0 ? (byte)transIdx : (byte)0);
                    bw.Write((byte)0);
                }
            }

            bw.Write((byte)0x3B);
        }

        private static byte[] GetPixelData(Bitmap bmp)
        {
            var data = new byte[bmp.Width * bmp.Height];
            var rect = new Rectangle(0, 0, bmp.Width, bmp.Height);
            var bmpData = bmp.LockBits(rect, ImageLockMode.ReadOnly, PixelFormat.Format32bppArgb);

            unsafe
            {
                byte* ptr = (byte*)bmpData.Scan0;
                int stride = bmpData.Stride;

                for (int y = 0; y < bmp.Height; y++)
                {
                    for (int x = 0; x < bmp.Width; x++)
                    {
                        byte alpha = *(ptr + y * stride + x * 4 + 3);
                        data[y * bmp.Width + x] = (alpha > 128 ? (byte)255 : (byte)0);
                    }
                }
            }

            bmp.UnlockBits(bmpData);
            return data;
        }

        private static byte[] LzwCompress(byte[] data, int minCodeSize)
        {
            var output = new MemoryStream();
            int clearCode = 1 << minCodeSize;
            int endCode = clearCode + 1;
            var table = new Dictionary<string, int>();

            for (int i = 0; i < clearCode; i++)
            {
                table[char.ToString((char)i)] = i;
            }

            int nextCode = endCode + 1;
            int codeSize = minCodeSize + 1;
            var buffer = new List<int>();
            string current = "";

            output.WriteByte((byte)minCodeSize);

            buffer.Add(clearCode);
            WriteCodes(buffer, codeSize, output);

            foreach (byte b in data)
            {
                string combined = current + (char)b;
                if (table.ContainsKey(combined))
                {
                    current = combined;
                }
                else
                {
                    buffer.Add(table[current]);
                    table[combined] = nextCode++;

                    if (nextCode > (1 << codeSize))
                    {
                        codeSize++;
                    }

                    if (nextCode >= 4096)
                    {
                        WriteCodes(buffer, codeSize, output);
                        buffer.Clear();
                        buffer.Add(clearCode);
                        WriteCodes(buffer, codeSize, output);

                        table.Clear();
                        for (int j = 0; j < clearCode; j++)
                        {
                            table[char.ToString((char)j)] = j;
                        }
                        nextCode = endCode + 1;
                        codeSize = minCodeSize + 1;
                    }

                    current = ((char)b).ToString();
                }
            }

            if (!string.IsNullOrEmpty(current))
            {
                buffer.Add(table[current]);
            }

            buffer.Add(endCode);
            WriteCodes(buffer, codeSize, output);

            return output.ToArray();
        }

        private static void WriteCodes(List<int> codes, int codeSize, Stream output)
        {
            int bitBuffer = 0;
            int bitsInBuffer = 0;
            var bytes = new List<byte>();

            foreach (int code in codes)
            {
                bitBuffer |= (code << bitsInBuffer);
                bitsInBuffer += codeSize;

                while (bitsInBuffer >= 8)
                {
                    bytes.Add((byte)(bitBuffer & 0xFF));
                    bitBuffer >>= 8;
                    bitsInBuffer -= 8;
                }
            }

            if (bitsInBuffer > 0)
            {
                bytes.Add((byte)(bitBuffer & 0xFF));
            }

            byte[] subBlockData = bytes.ToArray();
            int offset = 0;

            while (offset < subBlockData.Length)
            {
                int blockSize = Math.Min(255, subBlockData.Length - offset);
                output.Write(new byte[] { (byte)blockSize }, 0, 1);
                output.Write(subBlockData, offset, blockSize);
                offset += blockSize;
            }
        }
    }
}
