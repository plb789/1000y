using System;
using System.IO;

namespace MillenniumResEditor.Utils
{
    /// <summary>
    /// 二进制读写工具（统一小端序，匹配千年资源格式）
    /// </summary>
    public static class BinaryUtil
    {
        /// <summary>
        /// 读取 UInt16 小端
        /// </summary>
        public static ushort ReadUInt16LE(BinaryReader reader)
        {
            byte[] buf = reader.ReadBytes(2);
            if (BitConverter.IsLittleEndian)
                return BitConverter.ToUInt16(buf, 0);
            Array.Reverse(buf);
            return BitConverter.ToUInt16(buf, 0);
        }

        /// <summary>
        /// 写入 UInt16 小端
        /// </summary>
        public static void WriteUInt16LE(BinaryWriter writer, ushort value)
        {
            byte[] buf = BitConverter.GetBytes(value);
            if (!BitConverter.IsLittleEndian)
                Array.Reverse(buf);
            writer.Write(buf);
        }
    }
}