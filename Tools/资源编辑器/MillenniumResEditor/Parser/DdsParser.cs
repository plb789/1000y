using System;
using System.IO;
using System.Drawing;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 简易 DDS(DXT1) 解析 仅用于预览
    /// </summary>
    public class DdsParser
    {
        public Bitmap LoadDds(string filePath)
        {
            // 简易实现：Windows GDI 直接加载DDS（系统自带解码器）
            // 如需纯二进制解析，可扩展前端同款DXT1解码逻辑
            return new Bitmap(filePath);
        }
    }
}