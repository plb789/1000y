using System;
using System.Drawing;
using System.Windows.Forms;
using MillenniumResEditor.Model;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor
{
    public partial class FormSprViewer : Form
    {
        private readonly SprParser _sprParser;
        private readonly PalParser _palParser;
        private int _curFrame = 0;
        private readonly Timer _animTimer;

        public FormSprViewer(string sprPath)
        {
            InitializeComponent();
            _sprParser = new SprParser();
            _palParser = new PalParser();
            _animTimer = new Timer { Interval = 100 };
            _animTimer.Tick += AnimTick;

            // 同目录加载pal
            string palPath = System.IO.Path.ChangeExtension(sprPath, ".pal");
            if (System.IO.File.Exists(palPath))
                _palParser.Load(palPath);

            _sprParser.Load(sprPath);
            Text = $"SPR预览 - {System.IO.Path.GetFileName(sprPath)}";
            lblFrameCount.Text = $"总帧数：{_sprParser.Frames.Count}";

            _animTimer.Start();
        }

        private void AnimTick(object sender, EventArgs e)
        {
            if (_curFrame >= _sprParser.Frames.Count)
                _curFrame = 0;

            var frame = _sprParser.Frames[_curFrame];
            var bmp = DrawFrame(frame);
            picSpr.Image = bmp;
            lblCurFrame.Text = $"当前帧：{_curFrame + 1}";
            _curFrame++;
        }

        // 绘制单帧
        private Bitmap DrawFrame(SprFrame frame)
        {
            var bmp = new Bitmap(frame.Width, frame.Height);
            for (int y = 0; y < frame.Height; y++)
            {
                for (int x = 0; x < frame.Width; x++)
                {
                    int idx = y * frame.Width + x;
                    byte palIdx = frame.PixelData[idx];
                    var color = _palParser.Colors[palIdx].ToColor();
                    // 索引0设为透明
                    if (palIdx == 0) color = Color.Transparent;
                    bmp.SetPixel(x, y, color);
                }
            }
            return bmp;
        }

        private void FormSprViewer_FormClosed(object sender, FormClosedEventArgs e)
        {
            _animTimer.Stop();
        }
    }
}