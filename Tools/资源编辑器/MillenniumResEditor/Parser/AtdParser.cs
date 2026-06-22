using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using MillenniumResEditor.Model;

namespace MillenniumResEditor.Parser
{
    /// <summary>
    /// 千年 .atd 动画定义文件解析器
    ///
    /// 格式规范（来自原版 Delphi 源码 StrDb.pas + AtzCls.pas）：
    ///
    /// 文件由若干 255 字节的 TByteString 记录组成，每个记录经过 InverseByteString 加密
    /// （每个字节的高4位与低4位交换：b = (b and $F0 shr 4) + (b and $0F shl 4)）
    ///
    /// 解密后为 CSV 格式文本：
    ///   第1行：字段名（Name,Action,Direction,Frame,FrameTime,AF,AFpx,AFpy,BF,BFpx,BFpy,...）
    ///   后续行：每行一条动画记录，第一字段为记录名（1,2,3...）
    ///
    /// 字段含义：
    ///   Name      : 动画名称（通常与记录索引相同）
    ///   Action    : TURN/MOVE/MOVE1-5/HIT/HIT1-9/DIE/SEATDOWN/STANDUP/STRUCTED/HELLO
    ///   Direction : DR_0~DR_7（8方向，0=上，顺时针）
    ///   Frame     : 帧数
    ///   FrameTime : 每帧时长（毫秒）
    ///   AF,AFpx,AFpy : 第1帧的图像索引 + X偏移 + Y偏移
    ///   BF,BFpx,BFpy : 第2帧的图像索引 + X偏移 + Y偏移
    ///   ...（最多到 ZF，26帧）
    /// </summary>
    public class AtdParser
    {
        /// <summary>解析出的所有动画记录</summary>
        public List<AtdAnimation> Animations { get; } = new List<AtdAnimation>();

        /// <summary>字段名列表</summary>
        public List<string> Fields { get; } = new List<string>();

        /// <summary>从文件加载 .atd</summary>
        public void Load(string filePath)
        {
            using var fs = new FileStream(filePath, FileMode.Open, FileAccess.Read);
            LoadFromStream(fs);
        }

        /// <summary>从内存流加载（用于从 PGK 包中提取后直接解析）</summary>
        public void LoadFromStream(Stream stream)
        {
            Animations.Clear();
            Fields.Clear();

            // 1. 读取所有 255 字节记录并解密
            const int RecordSize = 255;
            List<string> records = new List<string>();

            stream.Position = 0;
            byte[] buffer = new byte[RecordSize];

            while (stream.Position + RecordSize <= stream.Length)
            {
                int read = stream.Read(buffer, 0, RecordSize);
                if (read < RecordSize) break;

                // InverseByteString：高低4位交换
                for (int i = 0; i < RecordSize; i++)
                {
                    byte b = buffer[i];
                    buffer[i] = (byte)(((b & 0xF0) >> 4) | ((b & 0x0F) << 4));
                }

                // 第0字节是字符串长度，之后是字符串内容
                int len = buffer[0];
                if (len > 0 && len <= RecordSize - 1)
                {
                    string str = Encoding.Default.GetString(buffer, 1, len);
                    records.Add(str);
                }
                else
                {
                    records.Add(string.Empty);
                }
            }

            // 2. 移除空记录和以逗号开头的记录
            for (int i = records.Count - 1; i >= 0; i--)
            {
                string s = records[i];
                if (string.IsNullOrEmpty(s) || s[0] == ',')
                    records.RemoveAt(i);
            }

            if (records.Count == 0) return;

            // 3. 第一行是字段名
            string[] fieldNames = records[0].Split(',');
            foreach (string f in fieldNames)
                Fields.Add(f.Trim());

            // 4. 后续每行是一条动画记录
            for (int i = 1; i < records.Count; i++)
            {
                string[] values = records[i].Split(',');
                if (values.Length < 5) continue;

                var anim = new AtdAnimation
                {
                    Name = values[0].Trim()
                };

                // 解析各字段
                for (int j = 1; j < values.Length && j < Fields.Count; j++)
                {
                    string fieldName = Fields[j];
                    string fieldValue = values[j].Trim();

                    switch (fieldName)
                    {
                        case "Action":
                            anim.Action = ParseAction(fieldValue);
                            anim.ActionName = fieldValue;
                            break;
                        case "Direction":
                            anim.Direction = ParseDirection(fieldValue);
                            anim.DirectionName = fieldValue;
                            break;
                        case "Frame":
                            anim.Frame = ParseInt(fieldValue);
                            break;
                        case "FrameTime":
                            anim.FrameTime = ParseInt(fieldValue);
                            break;
                    }
                }

                // 解析每帧信息（AF,AFpx,AFpy,BF,BFpx,BFpy,...）
                for (int f = 0; f < anim.Frame && f < 26; f++)
                {
                    char c = (char)('A' + f);
                    int idxF = Fields.IndexOf($"{c}F");
                    int idxPx = Fields.IndexOf($"{c}px");
                    int idxPy = Fields.IndexOf($"{c}py");

                    var frame = new AtdFrameInfo();
                    if (idxF >= 0 && idxF < values.Length)
                        frame.ImageIndex = ParseInt(values[idxF].Trim());
                    if (idxPx >= 0 && idxPx < values.Length)
                        frame.OffsetX = ParseInt(values[idxPx].Trim());
                    if (idxPy >= 0 && idxPy < values.Length)
                        frame.OffsetY = ParseInt(values[idxPy].Trim());

                    anim.Frames.Add(frame);
                }

                Animations.Add(anim);
            }
        }

        /// <summary>根据动作名称和方向查找动画</summary>
        public AtdAnimation FindAnimation(AnimAction action, int direction)
        {
            foreach (var anim in Animations)
            {
                if (anim.Action == action && anim.Direction == direction)
                    return anim;
            }
            return null;
        }

        private static AnimAction ParseAction(string str)
        {
            return str.ToUpperInvariant() switch
            {
                "TURN" => AnimAction.Turn,
                "TURNNING" => AnimAction.Turnning,
                "MOVE" => AnimAction.Move,
                "MOVE1" => AnimAction.Move1,
                "MOVE2" => AnimAction.Move2,
                "MOVE3" => AnimAction.Move3,
                "MOVE4" => AnimAction.Move4,
                "MOVE5" => AnimAction.Move5,
                "HIT" => AnimAction.Hit,
                "HIT1" => AnimAction.Hit1,
                "HIT2" => AnimAction.Hit2,
                "HIT3" => AnimAction.Hit3,
                "HIT4" => AnimAction.Hit4,
                "HIT5" => AnimAction.Hit5,
                "HIT6" => AnimAction.Hit6,
                "HIT7" => AnimAction.Hit7,
                "HIT8" => AnimAction.Hit8,
                "HIT9" => AnimAction.Hit9,
                "DIE" => AnimAction.Die,
                "STRUCTED" => AnimAction.Structed,
                "SEATDOWN" => AnimAction.SeatDown,
                "STANDUP" => AnimAction.StandUp,
                "HELLO" => AnimAction.Hello,
                _ => AnimAction.Unknown,
            };
        }

        private static int ParseDirection(string str)
        {
            // DR_0 ~ DR_7
            if (str.StartsWith("DR_") && int.TryParse(str.Substring(3), out int d))
                return d;
            return -1;
        }

        private static int ParseInt(string str)
        {
            int.TryParse(str, out int v);
            return v;
        }
    }
}
