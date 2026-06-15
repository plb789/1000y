using System.Collections.Generic;
using System.Drawing;

namespace UIDesigner
{
    /// <summary>千年游戏UI模板管理器</summary>
    public static class TemplateManager
    {
        /// <summary>模板名称列表</summary>
        public static List<string> GetTemplateNames()
        {
            return new List<string>
            {
                "登录窗口",
                "角色属性面板",
                "背包面板",
                "聊天窗口",
                "门派面板",
                "商城面板"
            };
        }

        /// <summary>加载指定模板控件集合</summary>
        public static List<ControlItem> LoadTemplate(string templateName)
        {
            return templateName switch
            {
                "登录窗口" => CreateLoginTemplate(),
                "角色属性面板" => CreateRoleTemplate(),
                "背包面板" => CreateBagTemplate(),
                "聊天窗口" => CreateChatTemplate(),
                "门派面板" => CreateFactionTemplate(),
                "商城面板" => CreateShopTemplate(),
                _ => new List<ControlItem>()
            };
        }

        #region 各个模板定义
        // 1. 登录窗口
        private static List<ControlItem> CreateLoginTemplate()
        {
            var list = new List<ControlItem>();
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Panel,
                CtrlId = "login_panel",
                X = 200, Y = 100, Width = 400, Height = 300,
                Radius = 12, BorderWidth = 2, BgColor = Color.FromArgb(60, 40, 20)
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Label,
                CtrlId = "login_title",
                Text = "千年江湖",
                X = 320, Y = 120, Width = 160, Height = 40,
                FontSize = 20, FontColor = Color.Gold
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.TextBox,
                CtrlId = "input_user",
                Text = "请输入账号",
                X = 260, Y = 180, Width = 280, Height = 35, Radius = 6
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.TextBox,
                CtrlId = "input_pwd",
                Text = "请输入密码",
                X = 260, Y = 220, Width = 280, Height = 35, Radius = 6
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Button,
                CtrlId = "btn_login",
                Text = "登录游戏",
                X = 300, Y = 260, Width = 180, Height = 35,
                BgColor = Color.DarkRed, FontColor = Color.White, Radius = 8,
                ClickEvent = "UI.pop('main_ui')"
            });
            return list;
        }

        // 2. 角色面板
        private static List<ControlItem> CreateRoleTemplate()
        {
            var list = new List<ControlItem>();
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Panel,
                CtrlId = "role_panel",
                X = 20, Y = 20, Width = 350, Height = 400,
                Radius = 10, BgColor = Color.FromArgb(45, 35, 25)
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Label,
                CtrlId = "role_name",
                Text = "角色名称",
                X = 40, Y = 40, Width = 120, Height = 30, FontColor = Color.Gold
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Label,
                CtrlId = "role_hp",
                Text = "生命值：1000",
                X = 40, Y = 80, Width = 150, Height = 25
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Label,
                CtrlId = "role_mp",
                Text = "内力：800",
                X = 40, Y = 110, Width = 150, Height = 25
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Button,
                CtrlId = "btn_close_role",
                Text = "关闭",
                X = 280, Y = 20, Width = 60, Height = 25,
                ClickEvent = "UI.closePop('role_panel')"
            });
            return list;
        }

        // 3. 背包面板
        private static List<ControlItem> CreateBagTemplate()
        {
            var list = new List<ControlItem>();
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Panel,
                CtrlId = "bag_panel",
                X = 50, Y = 50, Width = 450, Height = 380, Radius = 8
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.ListBox,
                CtrlId = "bag_list",
                X = 70, Y = 80, Width = 410, Height = 280
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Button,
                CtrlId = "btn_bag_close",
                Text = "关闭背包",
                X = 380, Y = 30, Width = 80, Height = 25,
                ClickEvent = "UI.closePop('bag_panel')"
            });
            return list;
        }

        // 4. 聊天窗口
        private static List<ControlItem> CreateChatTemplate()
        {
            var list = new List<ControlItem>();
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Panel,
                CtrlId = "chat_panel",
                X = 10, Y = 450, Width = 500, Height = 140, Radius = 6
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.ListBox,
                CtrlId = "chat_msg",
                X = 30, Y = 470, Width = 460, Height = 90
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.TextBox,
                CtrlId = "chat_input",
                Text = "输入聊天内容...",
                X = 30, Y = 570, Width = 380, Height = 25
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Button,
                CtrlId = "chat_send",
                Text = "发送",
                X = 420, Y = 570, Width = 70, Height = 25
            });
            return list;
        }

        // 5. 门派面板
        private static List<ControlItem> CreateFactionTemplate()
        {
            var list = new List<ControlItem>();
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Panel,
                CtrlId = "faction_panel",
                X = 150, Y = 80, Width = 420, Height = 360, Radius = 10
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Label,
                CtrlId = "faction_title",
                Text = "门派信息",
                X = 170, Y = 100, Width = 120, Height = 35, FontSize = 16, FontColor = Color.Gold
            });
            return list;
        }

        // 6. 商城面板
        private static List<ControlItem> CreateShopTemplate()
        {
            var list = new List<ControlItem>();
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.Panel,
                CtrlId = "shop_panel",
                X = 100, Y = 60, Width = 480, Height = 420, Radius = 10
            });
            list.Add(new ControlItem
            {
                CtrlType = UiControlType.ListBox,
                CtrlId = "shop_goods",
                X = 120, Y = 90, Width = 440, Height = 300
            });
            return list;
        }
        #endregion
    }
}