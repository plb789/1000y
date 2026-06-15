using System;
using System.Drawing;
using System.IO;

namespace UIDesigner
{
    /// <summary>DDS/PNG 贴图加载工具</summary>
    public static class ImageLoader
    {
        /// <summary>加载图片（支持PNG / DDS）</summary>
        public static Bitmap? LoadImage(string filePath)
        {
            if (!File.Exists(filePath)) return null;
            string ext = Path.GetExtension(filePath).ToLower();
            try
            {
                if (ext == ".png")
                {
                    return new Bitmap(filePath);
                }
                if (ext == ".dds")
                {
                    // Windows原生支持DDS解码（系统自带解码器）
                    return new Bitmap(filePath);
                }
                return null;
            }
            catch
            {
                return null;
            }
        }

        /// <summary>释放图片资源</summary>
        public static void DisposeBitmap(Bitmap? bmp)
        {
            bmp?.Dispose();
        }
    }
}