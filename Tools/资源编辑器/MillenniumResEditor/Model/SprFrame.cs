namespace MillenniumResEditor.Model
{
    /// <summary>
    /// SPR 单帧数据
    /// </summary>
    public class SprFrame
    {
        public ushort Width { get; set; }
        public ushort Height { get; set; }
        public ushort OffsetX { get; set; }
        public ushort OffsetY { get; set; }
        public byte[] PixelData { get; set; }
    }
}