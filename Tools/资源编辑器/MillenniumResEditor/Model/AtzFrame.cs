namespace MillenniumResEditor.Model
{
    /// <summary>
    /// ATZ 单帧数据（TA2ImageLib 中的一帧）
    /// </summary>
    public class AtzFrame
    {
        /// <summary>帧宽度</summary>
        public ushort Width { get; set; }

        /// <summary>帧高度</summary>
        public ushort Height { get; set; }

        /// <summary>X偏移（绘制时相对角色锚点的偏移）</summary>
        public short OffsetX { get; set; }

        /// <summary>Y偏移</summary>
        public short OffsetY { get; set; }

        /// <summary>
        /// RGB555 像素数据（每像素2字节）
        /// 长度 = Width * Height * 2
        /// 仅在 Format != Png 时有效
        /// </summary>
        public byte[] PixelData { get; set; }

        /// <summary>
        /// PNG 原始数据（ATZ5 格式）
        /// 仅在 Format == Png 时有效
        /// </summary>
        public byte[] PngData { get; set; }

        /// <summary>帧格式</summary>
        public AtzFormat Format { get; set; } = AtzFormat.RGB555;
    }

    /// <summary>
    /// ATZ 文件格式版本
    /// </summary>
    public enum AtzFormat
    {
        /// <summary>ATZ0 调色板模式</summary>
        Palette,

        /// <summary>ATZ1 RGB555 直接模式</summary>
        RGB555,

        /// <summary>ATZ3 加密调色板模式</summary>
        PaletteEncrypted,

        /// <summary>ATZ4 加密 RGB555 模式</summary>
        RGB555Encrypted,

        /// <summary>ATZ5 PNG 压缩模式</summary>
        Png
    }
}
