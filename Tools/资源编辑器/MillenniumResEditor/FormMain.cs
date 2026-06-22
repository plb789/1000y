using System;
using System.Windows.Forms;

namespace MillenniumResEditor
{
    public partial class FormMain : Form
    {
        public FormMain()
        {
            InitializeComponent();
            Text = "千年江湖 - 资源编辑器 V1.0";
        }

        // 打开地图编辑器
        private void btnOpenMap_Click(object sender, EventArgs e)
        {
            using var openDlg = new OpenFileDialog
            {
                Filter = "千年地图(*.map)|*.map|所有文件(*.*)|*.*",
                Title = "选择千年 .map 地图文件"
            };
            if (openDlg.ShowDialog() == DialogResult.OK)
            {
                var mapForm = new FormMapEditor(openDlg.FileName);
                mapForm.Show();
            }
        }

        // 打开SPR动画预览
        private void btnOpenSpr_Click(object sender, EventArgs e)
        {
            using var openDlg = new OpenFileDialog
            {
                Filter = "SPR动画(*.spr)|*.spr|所有文件(*.*)|*.*",
                Title = "选择 .spr 动画文件"
            };
            if (openDlg.ShowDialog() == DialogResult.OK)
            {
                var sprForm = new FormSprViewer(openDlg.FileName);
                sprForm.Show();
            }
        }

        // 打开DDS贴图预览
        private void btnOpenDds_Click(object sender, EventArgs e)
        {
            using var openDlg = new OpenFileDialog
            {
                Filter = "DDS贴图(*.dds)|*.dds|所有文件(*.*)|*.*",
                Title = "选择 .dds 贴图文件"
            };
            if (openDlg.ShowDialog() == DialogResult.OK)
            {
                var ddsForm = new FormDdsViewer(openDlg.FileName);
                ddsForm.Show();
            }
        }

        // 打开ATZ动画预览与导出
        private void btnOpenAtz_Click(object sender, EventArgs e)
        {
            var atzForm = new FormAtzViewer();
            atzForm.Show();
        }

        // 打开EFT特效查看与导出
        private void btnOpenEft_Click(object sender, EventArgs e)
        {
            var eftForm = new FormEftViewer();
            eftForm.Show();
        }

        // 打开通用DAT解包查看器
        private void btnOpenDat_Click(object sender, EventArgs e)
        {
            var datForm = new FormDatViewer();
            datForm.Show();
        }

        // 打开MAP DAT包查看器
        private void btnOpenMapDat_Click(object sender, EventArgs e)
        {
            var mapDatForm = new FormMapDatViewer();
            mapDatForm.Show();
        }
    }
}