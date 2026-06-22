using System.Collections.Generic;

namespace MillenniumResEditor.Model
{
    /// <summary>
    /// ATD 动画定义（对应原版 TAniInfo）
    /// </summary>
    public class AtdAnimation
    {
        /// <summary>动画名称（记录索引，如 "1"、"2"）</summary>
        public string Name { get; set; }

        /// <summary>动作类型</summary>
        public AnimAction Action { get; set; }

        /// <summary>动作名称（原始字符串，如 "MOVE"、"HIT"）</summary>
        public string ActionName { get; set; }

        /// <summary>方向（0-7，DR_0~DR_7）</summary>
        public int Direction { get; set; }

        /// <summary>方向名称（原始字符串）</summary>
        public string DirectionName { get; set; }

        /// <summary>帧数</summary>
        public int Frame { get; set; }

        /// <summary>每帧时长（毫秒）</summary>
        public int FrameTime { get; set; }

        /// <summary>每帧的详细信息</summary>
        public List<AtdFrameInfo> Frames { get; set; } = new List<AtdFrameInfo>();
    }

    /// <summary>
    /// ATD 单帧信息
    /// </summary>
    public class AtdFrameInfo
    {
        /// <summary>图像索引（对应 .atz 文件中的第几帧）</summary>
        public int ImageIndex { get; set; }

        /// <summary>X偏移</summary>
        public int OffsetX { get; set; }

        /// <summary>Y偏移</summary>
        public int OffsetY { get; set; }
    }

    /// <summary>
    /// 动作类型枚举（对应原版 AM_* 常量）
    /// 来自 AtzCls.pas 的 TAnimater.LoadFromFile
    /// </summary>
    public enum AnimAction
    {
        Unknown = -1,

        // 转身
        Turn = 0,
        Turn1,
        Turn2,
        Turn3,
        Turn4,
        Turn5,
        Turn6,
        Turn7,
        Turn8,
        Turn9,

        // 转身中
        Turnning,
        Turnning1,
        Turnning2,
        Turnning3,
        Turnning4,
        Turnning5,
        Turnning6,
        Turnning7,
        Turnning8,
        Turnning9,

        // 移动
        Move,
        Move1,
        Move2,
        Move3,
        Move4,
        Move5,
        Move6,
        Move7,
        Move8,
        Move9,

        // 攻击
        Hit,
        Hit1,
        Hit2,
        Hit3,
        Hit4,
        Hit5,
        Hit6,
        Hit7,
        Hit8,
        Hit9,

        // 其他
        Die,
        Structed,
        SeatDown,
        StandUp,
        Hello
    }
}
