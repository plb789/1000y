using System;
using System.Drawing;
using System.IO;
using System.Windows.Forms;
using MillenniumResEditor.Converter;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor
{
    /// <summary>
    /// 批量角色库导出界面
    ///
    /// 功能：
    ///   1. 选择 PGK 包或独立文件目录
    ///   2. 设置导出参数（最大身体ID、ATD索引范围等）
    ///   3. 执行批量导出，显示实时进度
    ///   4. 查看导出日志和统计信息
    /// </summary>
    public partial class FormBatchExport : Form
    {
        private BatchCharacterExporter _exporter;

        public FormBatchExport()
        {
            InitializeComponent();
        }

        // ==================== 事件处理 ====================

        private void CmbMode_SelectedIndexChanged(object sender, EventArgs e)
        {
            bool isPgk = (cmbMode.SelectedIndex == 0);

            // PGK 模式控件
            lblPgk1.Visible = isPgk;
            txtSpritePgk.Visible = isPgk;
            btnBrowsePgk.Visible = isPgk;
            lblPgk2.Visible = isPgk;
            txtSysPgk.Visible = isPgk;
            btnBrowseSysPgk.Visible = isPgk;

            // 独立文件模式控件
            lblAtzDir.Visible = !isPgk;
            txtAtzDir.Visible = !isPgk;
            btnBrowseAtzDir.Visible = !isPgk;
            lblAtdDir.Visible = !isPgk;
            txtAtdDir.Visible = !isPgk;
            btnBrowseAtdDir.Visible = !isPgk;
        }

        private void BtnBrowsePgk_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog { Filter = "PGK 包(*.pgk)|*.pgk|所有文件(*.*)|*.*", Title = "选择 sprite.pgk" };
            if (dlg.ShowDialog() == DialogResult.OK)
                txtSpritePgk.Text = dlg.FileName;
        }

        private void BtnBrowseSysPgk_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog { Filter = "PGK 包(*.pgk)|*.pgk|所有文件(*.*)|*.*", Title = "选择 sys.pgk" };
            if (dlg.ShowDialog() == DialogResult.OK)
                txtSysPgk.Text = dlg.FileName;
        }

        private void BtnBrowseAtzDir_Click(object sender, EventArgs e)
        {
            using var dlg = new FolderBrowserDialog { Description = "选择 ATZ 文件目录" };
            if (dlg.ShowDialog() == DialogResult.OK)
                txtAtzDir.Text = dlg.SelectedPath;
        }

        private void BtnBrowseAtdDir_Click(object sender, EventArgs e)
        {
            using var dlg = new FolderBrowserDialog { Description = "选择 ATD 文件目录" };
            if (dlg.ShowDialog() == DialogResult.OK)
                txtAtdDir.Text = dlg.SelectedPath;
        }

        private void BtnBrowseOutput_Click(object sender, EventArgs e)
        {
            using var dlg = new FolderBrowserDialog { Description = "选择输出目录" };
            if (dlg.ShowDialog() == DialogResult.OK)
                txtOutputDir.Text = dlg.SelectedPath;
        }

        private void BtnStartExport_Click(object sender, EventArgs e)
        {
            if (string.IsNullOrWhiteSpace(txtOutputDir.Text))
            {
                MessageBox.Show("请选择输出目录！", "提示");
                return;
            }

            try
            {
                // 创建导出器
                bool isPgk = (cmbMode.SelectedIndex == 0);

                if (isPgk)
                {
                    if (!File.Exists(txtSpritePgk.Text))
                    {
                        MessageBox.Show("请先选择 sprite.pgk 文件！", "提示");
                        return;
                    }
                    if (!File.Exists(txtSysPgk.Text))
                    {
                        MessageBox.Show("请先选择 sys.pgk 文件！", "提示");
                        return;
                    }

                    var spritePgk = new PgkExtractor();
                    spritePgk.Load(txtSpritePgk.Text);

                    var sysPgk = new PgkExtractor();
                    sysPgk.Load(txtSysPgk.Text);

                    _exporter = new BatchCharacterExporter(spritePgk, sysPgk);
                }
                else
                {
                    if (!Directory.Exists(txtAtzDir.Text))
                    {
                        MessageBox.Show("ATZ 目录不存在！", "提示");
                        return;
                    }

                    _exporter = new BatchCharacterExporter(txtAtzDir.Text, txtAtdDir.Text);
                }

                _exporter.OutputBase = txtOutputDir.Text;
                _exporter.ProgressCallback = OnProgress;

                int maxBody = (int)numMaxBodyId.Value;
                int maxAtd = (int)numMaxAtdIndex.Value;

                rtbLog.Clear();
                lblStats.Text = "";
                prgProgress.Value = 0;
                btnStartExport.Enabled = false;

                // 执行导出
                _exporter.ExportAll(maxBody, maxAtd, chkDefaultEquip.Checked);

                // 显示日志和统计
                foreach (string line in _exporter.Log)
                {
                    AppendLog(line);
                }

                lblStats.Text = $"统计：成功 {_exporter.Statistics.ExportedCount} 个角色 | " +
                               $"失败 {_exporter.Statistics.FailedCount} 个 | " +
                               $"总帧数 {_exporter.Statistics.TotalFramesExported} 帧 | " +
                               $"耗时 {_exporter.Statistics.ElapsedSeconds:F1} 秒";

                MessageBox.Show($"批量导出完成！\n\n成功: {_exporter.Statistics.ExportedCount} 个角色\n失败: {_exporter.Statistics.FailedCount} 个",
                    "完成", MessageBoxButtons.OK, MessageBoxIcon.Information);
            }
            catch (Exception ex)
            {
                AppendLog($"\n[错误] {ex.Message}");
                MessageBox.Show($"导出失败：{ex.Message}", "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
            finally
            {
                btnStartExport.Enabled = true;
            }
        }

        private void OnProgress(string message, int current, int total)
        {
            if (InvokeRequired)
            {
                Invoke(new Action(() => OnProgress(message, current, total)));
                return;
            }

            lblProgress.Text = message;
            if (total > 0)
            {
                prgProgress.Maximum = total;
                prgProgress.Value = Math.Min(current, total);
            }
            Application.DoEvents(); // 保持UI响应
        }

        private void AppendLog(string line)
        {
            rtbLog.AppendText(line + Environment.NewLine);
            rtbLog.ScrollToCaret();
        }
    }
}
