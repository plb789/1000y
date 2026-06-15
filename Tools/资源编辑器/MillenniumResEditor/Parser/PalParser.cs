using System.IO;
using MillenniumResEditor.Model;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年 .pal 256色调色板解析
    /// 格式：256 * 3Byte(RGB)
    /// </summary>
    public class PalParser
    {
        public PalColor[] Colors { get; } = new PalColor[256];

        public void Load(string filePath)
        {
            using var fs = new FileStream(filePath, FileMode.Open, FileAccess.Read);
            using var br = new BinaryReader(fs);

            for (int i = 0; i < 256; i++)
            {
                Colors[i] = new PalColor
                {
                    R = br.ReadByte(),
                    G = br.ReadByte(),
                    B = br.ReadByte()
                };
            }
        }
    }
}