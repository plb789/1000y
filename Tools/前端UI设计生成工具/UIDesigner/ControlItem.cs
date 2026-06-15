using System.Drawing;

namespace UIDesigner
{
    public enum UiControlType
    {
        Label,      // 文本标签
        Button,     // 按钮
        TextBox,    // 输入框
        Panel,      // 面板/容器
        ListBox,    // 列表（聊天/背包）
        ImageBox    // 新增：图片控件
    }

    /// <summary>UI控件模型（扩展版）</summary>
    public class ControlItem
    {
        // 基础类型
        public UiControlType CtrlType { get; set; }
        public string CtrlId { get; set; } = "";
        public string Text { get; set; } = "";

        // 位置尺寸
        public int X { get; set; }
        public int Y { get; set; }
        public int Width { get; set; }
        public int Height { get; set; }

        // 样式扩展
        public Color BgColor { get; set; } = Color.White;
        public Color FontColor { get; set; } = Color.Black;
        public int FontSize { get; set; } = 12;
        public int BorderWidth { get; set; } = 1;    // 边框宽度
        public Color BorderColor { get; set; } = Color.Gray; // 边框颜色
        public int Radius { get; set; } = 0;         // 圆角像素
        public int Opacity { get; set; } = 255;      // 透明度 0-255

        // 贴图（DDS/PNG）
        public string ImagePath { get; set; } = "";
        public Bitmap? BgImage { get; set; }

        // 逻辑标记（代码生成用）
        public bool IsHideDefault { get; set; } = false; // 默认隐藏
        public string ClickEvent { get; set; } = "";     // 自定义点击逻辑

        // 选中状态
        public bool IsSelected { get; set; } = false;

        public Rectangle GetRect()
        {
            return new Rectangle(X, Y, Width, Height);
        }
    }
}