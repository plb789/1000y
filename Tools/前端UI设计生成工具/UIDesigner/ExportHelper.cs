using System;
using System.IO;

namespace UIDesigner
{
    /// <summary>
    /// 目录&文件导出工具（对齐项目UI架构）
    /// </summary>
    public static class ExportHelper
    {
        /// <summary>
        /// 导出分类目录（Window/Widget/ChatUI等）
        /// </summary>
        public enum UiFolderType
        {
            Window,     // 窗口
            Widget,     // 通用控件
            ChatUI,     // 聊天
            RoleUI,     // 角色
            ShopUI,     // 商店
            SocialUI    // 社交
        }

        /// <summary>
        /// 一键导出整套UI文件
        /// </summary>
        public static bool ExportAll(string rootPath, UiFolderType folder, string fileName, string html, string css, string js)
        {
            try
            {
                // 拼接目标路径
                string subDir = folder switch
                {
                    UiFolderType.Window => "UI/Window",
                    UiFolderType.Widget => "UI/Widget",
                    UiFolderType.ChatUI => "UI/ChatUI",
                    UiFolderType.RoleUI => "UI/RoleUI",
                    UiFolderType.ShopUI => "UI/ShopUI",
                    UiFolderType.SocialUI => "UI/SocialUI",
                    _ => "UI"
                };

                string fullDir = Path.Combine(rootPath, subDir);
                Directory.CreateDirectory(fullDir);

                // 写入文件
                File.WriteAllText(Path.Combine(fullDir, $"{fileName}.html"), html);
                File.WriteAllText(Path.Combine(fullDir, $"{fileName}.css"), css);
                File.WriteAllText(Path.Combine(fullDir, $"{fileName}.js"), js);
                return true;
            }
            catch
            {
                return false;
            }
        }
    }
}