using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年 PGK/DGK 资源包提取器
    ///
    /// 格式规范（来自原版 Delphi 源码 Client\v5ncl1000\PGK\FfilePgk.pas）：
    ///
    /// 文件整体布局：
    ///   [0-15]   : 16字节标志（4个Int32：34234243/34567443/22566788/22347788 或旧版其他值）
    ///   [16-84]  : TFILEHEAD（69字节，加密存储）
    ///   [85-...] : TFileListdata_file 数组（每个73字节，加密存储）
    ///   之后为各文件的原始数据（加密存储）
    ///
    /// TFILEHEAD（69字节，Delphi record）：
    ///   rname[65]     : 包名 ShortString（首字节=长度，固定"fpk"）
    ///   rfilecount(4) : 文件数量 Int32
    ///
    /// TFileListdata_file（73字节，Delphi record）：
    ///   rname[65]  : 文件名 ShortString（加密后的大写十六进制名如"EC889C466D2011C2266B.dat"）
    ///   radds(4)   : 数据偏移 Int32
    ///   rsize(4)   : 数据大小 Int32
    ///
    /// 加密方式（关键！）：
    ///   写入时调用 DeCryption(buf) = 每字节 ror 5（循环右移5位）
    ///   读取时调用 EnCryption(buf) = 每字节 rol 5（循环左移5位），恢复原文
    /// </summary>
    public class PgkExtractor
    {
        // XorKey（来自 pgktools/Unit1.pas，用于文件名加解密）
        private static readonly byte[] XorKey =
            { 0xA2, 0xB8, 0xAC, 0x68, 0x2C, 0x74, 0x4B, 0x92, 0x68, 0x2C, 0x64 };

        // 结构体大小常量（来自 FfilePgk.pas）
        private const int SizeOfTFileHead = 69;      // string[64](65) + integer(4) = 69
        private const int SizeOfFileListEntry = 73;   // string[64](65) + integer(4) + integer(4) = 73
        private const int MaxEntries = 65535;

        /// <summary>包内文件列表</summary>
        public List<PgkFileEntry> Files { get; } = new List<PgkFileEntry>();

        /// <summary>包文件路径</summary>
        public string PackagePath { get; private set; }

        /// <summary>从文件加载 PGK/DGK 包</summary>
        public void Load(string filePath)
        {
            PackagePath = filePath;
            Files.Clear();

            byte[] allData = File.ReadAllBytes(filePath);

            if (allData.Length < 16 + SizeOfTFileHead)
                throw new InvalidDataException("文件太小，不是有效的 PGK 包");

            // 1. 跳过16字节标志（位置 0-15）
            int pos = 16;

            // 2. 读取并解密 TFILEHEAD（69字节）
            if (pos + SizeOfTFileHead > allData.Length)
                throw new InvalidDataException("文件太小，无法读取文件头");

            byte[] headBytes = new byte[SizeOfTFileHead];
            Array.Copy(allData, pos, headBytes, 0, SizeOfTFileHead);
            // 关键修正：读取后用 Encryption(rol 5) 解密（原版 LoadFromFile_02 中 ReadBuffer 后调 EnCryption）
            Encryption(headBytes, SizeOfTFileHead);
            pos += SizeOfTFileHead;

            // 解析 ShortString 包名
            int nameLen = headBytes[0];
            string packageName = Encoding.Default.GetString(headBytes, 1, Math.Min(nameLen, 64));

            // 解析文件数量（小端序 Int32，偏移在 nameLen+1 之后？不，是固定偏移 65）
            int fileCount = BitConverter.ToInt32(headBytes, 65);

            if (fileCount <= 0 || fileCount > MaxEntries)
                throw new InvalidDataException($"无效的文件数量：{fileCount}");

            // 3. 验证剩余空间足够读取目录项
            int dirSizeNeeded = fileCount * SizeOfFileListEntry;
            if (pos + dirSizeNeeded > allData.Length)
                throw new InvalidDataException(
                    $"文件太小，需要 {pos + dirSizeNeeded} 字节读取目录，实际只有 {allData.Length} 字节");

            // 4. 批量读取所有目录项并解密
            byte[] dirBytes = new byte[dirSizeNeeded];
            Array.Copy(allData, pos, dirBytes, 0, dirSizeNeeded);
            // 关键修正：读取后用 Encryption(rol 5) 解密
            Encryption(dirBytes, dirSizeNeeded);

            // 5. 逐条解析目录项
            for (int i = 0; i < fileCount; i++)
            {
                int entryOffset = i * SizeOfFileListEntry;
                if (entryOffset + SizeOfFileListEntry > dirBytes.Length) break;

                int entryNameLen = dirBytes[entryOffset];
                entryNameLen = Math.Min(entryNameLen, 64); // 安全限制
                string entryName = Encoding.Default.GetString(dirBytes, entryOffset + 1, entryNameLen);
                int adds = BitConverter.ToInt32(dirBytes, entryOffset + 65);
                int size = BitConverter.ToInt32(dirBytes, entryOffset + 69);

                // 跳过无效条目
                if (size > 0 && adds >= 0 && adds + size <= allData.Length)
                {
                    Files.Add(new PgkFileEntry
                    {
                        Name = entryName,
                        Offset = adds,
                        Size = size
                    });
                }
            }
        }

        /// <summary>
        /// 提取指定文件到内存流
        /// </summary>
        /// <param name="fileName">加密后的文件名（如 "EC889C466D2011C2266B.dat"）</param>
        public MemoryStream Extract(string fileName)
        {
            PgkFileEntry entry = null;
            foreach (var f in Files)
            {
                if (string.Equals(f.Name, fileName, StringComparison.OrdinalIgnoreCase))
                {
                    entry = f;
                    break;
                }
            }
            if (entry == null) return null;

            byte[] allData = File.ReadAllBytes(PackagePath);
            if (entry.Offset + entry.Size > allData.Length)
                return null;

            byte[] data = new byte[entry.Size];
            Array.Copy(allData, entry.Offset, data, 0, entry.Size);

            // 关键修正：读取后用 Encryption(rol 5) 解密（原版 get() 方法中 ReadBuffer 后调 EnCryption）
            Encryption(data, data.Length);

            return new MemoryStream(data);
        }

        /// <summary>
        /// 提取文件并返回诊断信息（用于调试，同时测试三种解密方向）
        /// </summary>
        public (MemoryStream Stream, string DiagInfo) ExtractWithDiag(string fileName)
        {
            PgkFileEntry entry = null;
            foreach (var f in Files)
            {
                if (string.Equals(f.Name, fileName, StringComparison.OrdinalIgnoreCase))
                {
                    entry = f;
                    break;
                }
            }
            if (entry == null) return (null, $"文件未找到: {fileName}");

            byte[] allData = File.ReadAllBytes(PackagePath);
            var sb = new StringBuilder();
            sb.AppendLine($"包大小: {allData.Length} 字节");
            sb.AppendLine($"条目: {entry.Name} | 偏移={entry.Offset} | 大小={entry.Size}");

            if (entry.Offset + entry.Size > allData.Length)
                return (null, $"偏移越界: {entry.Offset}+{entry.Size} > {allData.Length}");

            // 读取原始数据
            byte[] rawData = new byte[entry.Size];
            Array.Copy(allData, entry.Offset, rawData, 0, entry.Size);

            int peekLen = Math.Min(32, rawData.Length);
            sb.AppendLine($"=== 原始数据前{peekLen}字节 ===");
            sb.AppendLine(BitConverter.ToString(rawData, 0, peekLen).Replace("-", " "));

            // 方向1：ROL 5（左移5位）
            byte[] d1 = new byte[rawData.Length]; Array.Copy(rawData, d1, rawData.Length);
            Encryption(d1, d1.Length);
            sb.AppendLine($"\n=== ROL5 解密后 ===");
            sb.AppendLine(BitConverter.ToString(d1, 0, peekLen).Replace("-", " "));
            if (d1.Length >= 20)
                sb.AppendLine($"Ident@16: '{Encoding.ASCII.GetString(d1,16,4)}'");

            // 方向2：ROR 5（右移5位）
            byte[] d2 = new byte[rawData.Length]; Array.Copy(rawData, d2, rawData.Length);
            Decryption(d2, d2.Length);
            sb.AppendLine($"\n=== ROR5 解密后 ===");
            sb.AppendLine(BitConverter.ToString(d2, 0, peekLen).Replace("-", " "));
            if (d2.Length >= 20)
                sb.AppendLine($"Ident@16: '{Encoding.ASCII.GetString(d2,16,4)}'");

            // 方向3：不解密
            sb.AppendLine($"\n=== 不解密 ===");
            if (rawData.Length >= 20)
                sb.AppendLine($"Ident@16: '{Encoding.ASCII.GetString(rawData,16,4)}'");

            return (new MemoryStream(d1), sb.ToString());
        }

        /// <summary>
        /// 提取所有文件到指定目录
        /// </summary>
        public void ExtractAll(string outputDir, IProgress<int> progress = null)
        {
            Directory.CreateDirectory(outputDir);
            for (int i = 0; i < Files.Count; i++)
            {
                var entry = Files[i];
                using var ms = Extract(entry.Name);
                if (ms == null) continue;

                string outPath = Path.Combine(outputDir, entry.RealName);
                File.WriteAllBytes(outPath, ms.ToArray());

                progress?.Report(i + 1);
            }
        }

        /// <summary>
        /// 解密文件名（Enc 的逆运算）
        /// 原版 Enc: 每字符 xor XorKey[j], j=(j+1)%11, 结果转两位十六进制大写
        /// Dec: 两位十六进制转字节, xor XorKey[j], 得到原始字符
        /// </summary>
        public static string DecryptFileName(string encName)
        {
            if (string.IsNullOrEmpty(encName)) return string.Empty;

            // 去掉 .dat 后缀
            string hex = encName;
            if (hex.EndsWith(".dat", StringComparison.OrdinalIgnoreCase))
                hex = hex.Substring(0, hex.Length - 4);

            var sb = new StringBuilder();
            int j = 0;
            for (int i = 0; i + 2 <= hex.Length; i += 2)
            {
                string hexByte = hex.Substring(i, 2);
                byte b = Convert.ToByte(hexByte, 16);
                b ^= XorKey[j];
                sb.Append((char)b);
                j = (j + 1) % 11;
            }
            return sb.ToString();
        }

        /// <summary>
        /// 加密文件名（原版 Enc 函数）
        /// </summary>
        public static string EncryptFileName(string plainName)
        {
            var sb = new StringBuilder();
            int j = 0;
            foreach (char c in plainName)
            {
                byte b = (byte)(c ^ XorKey[j]);
                sb.Append(b.ToString("X2"));
                j = (j + 1) % 11;
            }
            return sb.ToString();
        }

        /// <summary>
        /// 加密（循环左移5位，对应原版 EnCryption）
        /// 用于从 PGK 包读取数据后恢复原文
        /// 原理：写入时用了 DeCryption(ror 5)，所以读取时需用 EnCryption(rol 5) 还原
        /// </summary>
        public static void Encryption(byte[] buf, int size)
        {
            for (int i = 0; i < size; i++)
            {
                byte b = buf[i];
                buf[i] = (byte)((b << 5) | (b >> 3)); // rol 5
            }
        }

        /// <summary>
        /// 解密（循环右移5位，对应原版 DeCryption）
        /// 仅用于写入 PGK 包时的预处理
        /// </summary>
        private static void Decryption(byte[] buf, int size)
        {
            for (int i = 0; i < size; i++)
            {
                byte b = buf[i];
                buf[i] = (byte)((b >> 5) | (b << 3)); // ror 5
            }
        }
    }

    /// <summary>
    /// PGK 包内文件项
    /// </summary>
    public class PgkFileEntry
    {
        /// <summary>文件名（加密后的名称，如 "EC889C466D2011C2266B.dat"）</summary>
        public string Name { get; set; }

        /// <summary>数据在包中的偏移</summary>
        public int Offset { get; set; }

        /// <summary>数据大小</summary>
        public int Size { get; set; }

        /// <summary>解密后的真实文件名（PGK内直接存明文，无需解密）</summary>
        public string RealName => Name;
    }
}
