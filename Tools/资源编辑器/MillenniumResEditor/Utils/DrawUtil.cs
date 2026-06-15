using System;
using System.Drawing;

namespace MillenniumResEditor.Utils
{
    /// <summary>
    /// GDI+ 绘图通用工具类
    /// </summary>
    public static class DrawUtil
    {
        /// <summary>
        /// 创建纯色位图
        /// </summary>
        public static Bitmap CreateSolidBitmap(int width, int height, Color color)
        {
            var bmp = new Bitmap(width, height);
            using (var g = Graphics.FromImage(bmp))
            {
                g.Clear(color);
            }
            return bmp;
        }

        /// <summary>
        /// 绘制空心网格线
        /// </summary>
        public static void DrawGrid(Graphics g, int tileSize, int mapW, int mapH, Color lineColor)
        {
            using (var pen = new Pen(lineColor, 1))
            {
                // 垂直线
                for (int x = 0; x <= mapW; x += tileSize)
                {
                    g.DrawLine(pen, x, 0, x, mapH);
                }
                // 水平线
                for (int y = 0; y <= mapH; y += tileSize)
                {
                    g.DrawLine(pen, 0, y, mapW, y);
                }
            }
        }

        /// <summary>
        /// 安全释放GDI资源
        /// </summary>
        public static void SafeDispose(IDisposable obj)
        {
            obj?.Dispose();
        }
    }
}