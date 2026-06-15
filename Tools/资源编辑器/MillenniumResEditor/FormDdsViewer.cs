using System;
using System.Windows.Forms;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor
{
    public partial class FormDdsViewer : Form
    {
        public FormDdsViewer(string ddsPath)
        {
            InitializeComponent();
            var parser = new DdsParser();
            try
            {
                picDds.Image = parser.LoadDds(ddsPath);
                Text = $"DDS预览 - {System.IO.Path.GetFileName(ddsPath)}";
            }
            catch (Exception ex)
            {
                MessageBox.Show("加载DDS失败：" + ex.Message);
            }
        }
    }
}