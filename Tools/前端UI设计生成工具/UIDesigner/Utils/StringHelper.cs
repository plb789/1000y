namespace UIDesigner.Utils
{
    public static class StringHelper
    {
        /// <summary>
        /// 缩进格式化
        /// </summary>
        public static string Indent(string content, int space = 4)
        {
            string pad = new string(' ', space);
            return pad + content.Replace("\r\n", "\r\n" + pad);
        }
    }
}