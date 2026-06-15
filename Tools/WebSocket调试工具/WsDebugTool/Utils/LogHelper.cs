using System;
using System.Windows.Forms;

namespace WsDebugTool.Utils
{
    /// <summary>
    /// 日志输出工具
    /// </summary>
    public static class LogHelper
    {
        /// <summary>
        /// 追加日志到文本框
        /// </summary>
        public static void AppendLog(TextBox txt, string msg, bool isError = false)
        {
            string time = DateTime.Now.ToString("HH:mm:ss");
            string prefix = isError ? "[错误]" : "[正常]";
            string line = $"[{time}] {prefix} {msg}\r\n";

            if (txt.InvokeRequired)
            {
                txt.Invoke(new Action(() =>
                {
                    txt.AppendText(line);
                    txt.ScrollToCaret();
                }));
            }
            else
            {
                txt.AppendText(line);
                txt.ScrollToCaret();
            }
        }

        /// <summary>
        /// 清空日志
        /// </summary>
        public static void ClearLog(TextBox txt)
        {
            if (txt.InvokeRequired)
            {
                txt.Invoke(new Action(txt.Clear));
            }
            else
            {
                txt.Clear();
            }
        }
    }
}