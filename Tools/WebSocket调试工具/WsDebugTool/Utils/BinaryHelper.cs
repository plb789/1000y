using System;
using System.Linq;

namespace WsDebugTool.Utils
{
    /// <summary>
    /// 二进制工具、封包编解码（和Go服务端规则完全一致）
    /// 协议：[Cmd(2LE)][Len(2LE)][Body][Check(1)]
    /// </summary>
    public static class BinaryHelper
    {
        /// <summary>
        /// 小端序读取 ushort
        /// </summary>
        public static ushort ReadUInt16LE(byte[] data, int offset)
        {
            byte[] buf = new byte[2];
            Array.Copy(data, offset, buf, 0, 2);
            if (!BitConverter.IsLittleEndian)
                Array.Reverse(buf);
            return BitConverter.ToUInt16(buf, 0);
        }

        /// <summary>
        /// 小端序写入 ushort
        /// </summary>
        public static byte[] WriteUInt16LE(ushort value)
        {
            byte[] buf = BitConverter.GetBytes(value);
            if (!BitConverter.IsLittleEndian)
                Array.Reverse(buf);
            return buf;
        }

        /// <summary>
        /// 计算累加校验码（全字节求和 & 0xFF）
        /// </summary>
        public static byte CalcCheckSum(byte[] data)
        {
            int sum = data.Sum(b => b);
            return (byte)(sum & 0xFF);
        }

        /// <summary>
        /// 组装完整封包
        /// </summary>
        /// <param name="cmd">协议号</param>
        /// <param name="body">消息体</param>
        /// <returns>完整二进制封包</returns>
        public static byte[] EncodePacket(ushort cmd, byte[] body)
        {
            body ??= Array.Empty<byte>();
            int bodyLen = body.Length;
            int totalLen = 4 + bodyLen + 1;

            byte[] packet = new byte[totalLen];
            // 写入协议号 2字节 小端
            byte[] cmdBuf = WriteUInt16LE(cmd);
            Array.Copy(cmdBuf, 0, packet, 0, 2);
            // 写入长度 2字节 小端
            byte[] lenBuf = WriteUInt16LE((ushort)bodyLen);
            Array.Copy(lenBuf, 0, packet, 2, 2);
            // 写入消息体
            Array.Copy(body, 0, packet, 4, bodyLen);
            // 计算并写入校验码
            packet[totalLen - 1] = CalcCheckSum(packet.Take(totalLen - 1).ToArray());

            return packet;
        }

        /// <summary>
        /// 解析封包
        /// </summary>
        /// <param name="packet">完整封包</param>
        /// <param name="cmd">输出协议号</param>
        /// <param name="body">输出消息体</param>
        /// <returns>校验是否通过</returns>
        public static bool DecodePacket(byte[] packet, out ushort cmd, out byte[] body)
        {
            cmd = 0;
            body = Array.Empty<byte>();
            if (packet.Length < 5)
                return false;

            // 校验码验证
            byte realCheck = packet[packet.Length - 1];
            byte calcCheck = CalcCheckSum(packet.Take(packet.Length - 1).ToArray());
            if (realCheck != calcCheck)
                return false;

            // 解析协议号 + 长度
            cmd = ReadUInt16LE(packet, 0);
            ushort bodyLen = ReadUInt16LE(packet, 2);
            if (4 + bodyLen > packet.Length - 1)
                return false;

            body = new byte[bodyLen];
            Array.Copy(packet, 4, body, 0, bodyLen);
            return true;
        }

        /// <summary>
        /// 字节数组转十六进制字符串（方便查看）
        /// </summary>
        public static string BytesToHex(byte[] data)
        {
            if (data == null || data.Length == 0)
                return "";
            return BitConverter.ToString(data).Replace("-", " ");
        }
    }
}