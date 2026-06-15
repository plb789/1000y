using System.Drawing;

namespace MillenniumResEditor.Model
{
    /// <summary>
    /// PAL 单颜色
    /// </summary>
    public class PalColor
    {
        public byte R { get; set; }
        public byte G { get; set; }
        public byte B { get; set; }

        public Color ToColor()
        {
            return Color.FromArgb(R, G, B);
        }
    }
}