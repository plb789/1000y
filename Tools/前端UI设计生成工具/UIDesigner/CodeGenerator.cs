using System.Collections.Generic;
using System.Drawing;
using System.Text;
using UIDesigner.Utils;

namespace UIDesigner
{
    public class CodeGenerator
    {
        private readonly List<ControlItem> _controlList;

        public CodeGenerator(List<ControlItem> list)
        {
            _controlList = list;
        }

        public string GenerateHtml(string pageTitle = "游戏UI页面")
        {
            StringBuilder html = new StringBuilder();
            html.AppendLine("<!DOCTYPE html>");
            html.AppendLine("<html lang=\"zh-CN\">");
            html.AppendLine("<head>");
            html.AppendLine(StringHelper.Indent($"<meta charset=\"UTF-8\">", 4));
            html.AppendLine(StringHelper.Indent($"<title>{pageTitle}</title>", 4));
            html.AppendLine(StringHelper.Indent("<link rel=\"stylesheet\" href=\"style.css\">", 4));
            html.AppendLine("</head>");
            html.AppendLine("<body>");

            foreach (var item in _controlList)
            {
                html.AppendLine(StringHelper.Indent(BuildControlHtml(item), 4));
            }

            html.AppendLine("</body>");
            html.AppendLine("</html>");
            return html.ToString();
        }

        public string GenerateCss()
        {
            StringBuilder css = new StringBuilder();
            css.AppendLine("/* 千年江湖UI样式 - 自动生成 */");
            css.AppendLine("body { margin: 0; padding: 0; background: #222; }");
            css.AppendLine();

            foreach (var item in _controlList)
            {
                if (string.IsNullOrEmpty(item.CtrlId)) continue;
                css.AppendLine($"#{item.CtrlId} {{");
                css.AppendLine(StringHelper.Indent($"position: absolute;", 4));
                css.AppendLine(StringHelper.Indent($"left: {item.X}px;", 4));
                css.AppendLine(StringHelper.Indent($"top: {item.Y}px;", 4));
                css.AppendLine(StringHelper.Indent($"width: {item.Width}px;", 4));
                css.AppendLine(StringHelper.Indent($"height: {item.Height}px;", 4));
                css.AppendLine(StringHelper.Indent($"background-color: {ColorToHex(item.BgColor)};", 4));
                css.AppendLine(StringHelper.Indent($"color: {ColorToHex(item.FontColor)};", 4));
                css.AppendLine(StringHelper.Indent($"font-size: {item.FontSize}px;", 4));
                css.AppendLine(StringHelper.Indent($"border: {item.BorderWidth}px solid {ColorToHex(item.BorderColor)};", 4));
                css.AppendLine(StringHelper.Indent($"border-radius: {item.Radius}px;", 4));
                css.AppendLine(StringHelper.Indent($"opacity: {item.Opacity / 255f:F2};", 4));

                // 背景贴图
                if (!string.IsNullOrEmpty(item.ImagePath))
                {
                    css.AppendLine(StringHelper.Indent($"background-image: url('{item.ImagePath}');", 4));
                    css.AppendLine(StringHelper.Indent($"background-size: 100% 100%;", 4));
                    css.AppendLine(StringHelper.Indent($"background-repeat: no-repeat;", 4));
                }
                // 默认隐藏
                if (item.IsHideDefault)
                {
                    css.AppendLine(StringHelper.Indent($"display: none;", 4));
                }
                css.AppendLine("}");
                css.AppendLine();
            }
            return css.ToString();
        }

        /// <summary>增强JS：预制弹窗、显隐、跳转通用逻辑</summary>
        public string GenerateJs()
        {
            StringBuilder js = new StringBuilder();
            js.AppendLine("// 千年江湖UI通用逻辑 - 自动生成");
            js.AppendLine("// 通用工具函数");
            js.AppendLine("const UI = {");
            js.AppendLine("  // 显示控件");
            js.AppendLine("  show(id){const el=document.getElementById(id);if(el)el.style.display='block';},");
            js.AppendLine("  // 隐藏控件");
            js.AppendLine("  hide(id){const el=document.getElementById(id);if(el)el.style.display='none';},");
            js.AppendLine("  // 弹窗");
            js.AppendLine("  pop(id){this.show(id);},");
            js.AppendLine("  // 关闭弹窗");
            js.AppendLine("  closePop(id){this.hide(id);},");
            js.AppendLine("  // 页面跳转");
            js.AppendLine("  jump(url){window.location.href=url;}");
            js.AppendLine("};");
            js.AppendLine();
            js.AppendLine("window.onload = function(){");
            js.AppendLine(StringHelper.Indent("// 控件事件绑定", 4));

            foreach (var item in _controlList)
            {
                if (string.IsNullOrEmpty(item.CtrlId)) continue;
                if (item.CtrlType == UiControlType.Button)
                {
                    string eventCode = string.IsNullOrEmpty(item.ClickEvent)
                        ? "// 按钮自定义逻辑"
                        : item.ClickEvent;
                    js.AppendLine(StringHelper.Indent("document.getElementById('" + item.CtrlId + "').addEventListener('click',function(){", 4));
                    js.AppendLine(StringHelper.Indent(eventCode, 8));
                    js.AppendLine(StringHelper.Indent("});", 4));
                }
            }
            js.AppendLine("};");
            return js.ToString();
        }

        #region 内部方法
        private string BuildControlHtml(ControlItem item)
        {
            string id = string.IsNullOrEmpty(item.CtrlId) ? "" : $"id=\"{item.CtrlId}\"";
            return item.CtrlType switch
            {
                UiControlType.Label => $"<div {id}>{item.Text}</div>",
                UiControlType.Button => $"<button {id}>{item.Text}</button>",
                UiControlType.TextBox => $"<input type=\"text\" {id} placeholder=\"{item.Text}\">",
                UiControlType.Panel => $"<div {id} class=\"ui-panel\">{item.Text}</div>",
                UiControlType.ListBox => $"<ul {id}></ul>",
                UiControlType.ImageBox => $"<div {id}></div>",
                _ => $"<div {id}>{item.Text}</div>"
            };
        }

        private string ColorToHex(Color color)
        {
            return $"#{color.R:X2}{color.G:X2}{color.B:X2}";
        }
        #endregion
    }
}