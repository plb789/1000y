using System;
using System.Drawing;
using System.Drawing.Drawing2D;
using System.Drawing.Imaging;

namespace UIDesigner.Utils
{
    public static class DrawHelper
    {
        /// <summary>绘制选中虚线框</summary>
        public static void DrawSelectRect(Graphics g, Rectangle rect)
        {
            using (Pen pen = new Pen(Color.Blue, 1) { DashStyle = DashStyle.Dash })
            {
                g.DrawRectangle(pen, rect);
            }
        }

        /// <summary>绘制带样式控件（圆角、贴图、透明度、边框）</summary>
        public static void DrawControl(Graphics g, ControlItem item)
        {
            var rect = item.GetRect();
            g.SmoothingMode = SmoothingMode.AntiAlias;

            // 1. 圆角路径
            GraphicsPath path = GetRoundRectPath(rect, item.Radius);
            g.SetClip(path);

            // 2. 绘制背景贴图 / 纯色背景
            if (item.BgImage != null)
            {
                float alpha = item.Opacity / 255f;
                using (var ia = new ImageAttributes())
                {
                    ia.SetColorMatrix(new ColorMatrix
                    {
                        Matrix33 = alpha
                    });
                    g.DrawImage(item.BgImage, rect, 0, 0, item.BgImage.Width, item.BgImage.Height, GraphicsUnit.Pixel, ia);
                }
            }
            else
            {
                using (Brush brush = new SolidBrush(Color.FromArgb(item.Opacity, item.BgColor)))
                {
                    g.FillPath(brush, path);
                }
            }

            // 3. 绘制边框
            if (item.BorderWidth > 0)
            {
                using (Pen pen = new Pen(item.BorderColor, item.BorderWidth))
                {
                    g.DrawPath(pen, path);
                }
            }

            // 4. 绘制文本
            if (!string.IsNullOrEmpty(item.Text))
            {
                using (Brush fBrush = new SolidBrush(Color.FromArgb(item.Opacity, item.FontColor)))
                using (Font font = new Font("微软雅黑", item.FontSize))
                {
                    var format = new StringFormat
                    {
                        Alignment = StringAlignment.Center,
                        LineAlignment = StringAlignment.Center
                    };
                    g.DrawString(item.Text, font, fBrush, rect, format);
                }
            }

            g.ResetClip();
            path.Dispose();

            // 选中框
            if (item.IsSelected)
            {
                DrawSelectRect(g, rect);
            }
        }

        /// <summary>生成圆角矩形路径</summary>
        private static GraphicsPath GetRoundRectPath(Rectangle rect, int radius)
        {
            var path = new GraphicsPath();
            if (radius <= 0)
            {
                path.AddRectangle(rect);
                return path;
            }
            int r = Math.Min(radius, rect.Width / 2);
            path.AddArc(rect.X, rect.Y, r, r, 180, 90);
            path.AddArc(rect.Right - r, rect.Y, r, r, 270, 90);
            path.AddArc(rect.Right - r, rect.Bottom - r, r, r, 0, 90);
            path.AddArc(rect.X, rect.Bottom - r, r, r, 90, 90);
            path.CloseAllFigures();
            return path;
        }
    }
}