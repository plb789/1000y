using System;
using System.Collections.Generic;
using System.IO;
using System.Windows.Forms;
using MillenniumResEditor.Converter;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor
{
    /// <summary>
    /// 批量合成角色动画导出窗体
    /// 用户指定 sprite.pgk / sys.pgk 路径和角色参数，
    /// 自动合成所有动作和方向的 PNG 序列帧，并生成 manifest.json
    /// </summary>
    public partial class FormCharacterExporter : Form
    {
        public FormCharacterExporter()
        {
            InitializeComponent();
        }

        // ==================== 事件处理 ====================

        private void btnBrowseSprite_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog { Filter = "PGK包(*.pgk)|*.pgk|所有文件(*.*)|*.*", Title = "选择 sprite.pgk" };
            if (dlg.ShowDialog() == DialogResult.OK) txtSpritePgk.Text = dlg.FileName;
        }

        private void btnBrowseSys_Click(object sender, EventArgs e)
        {
            using var dlg = new OpenFileDialog { Filter = "PGK包(*.pgk)|*.pgk|所有文件(*.*)|*.*", Title = "选择 sys.pgk" };
            if (dlg.ShowDialog() == DialogResult.OK) txtSysPgk.Text = dlg.FileName;
        }

        private void btnBrowseOutput_Click(object sender, EventArgs e)
        {
            using var dlg = new FolderBrowserDialog { Description = "选择输出目录（将创建 characters 子目录）" };
            if (dlg.ShowDialog() == DialogResult.OK) txtOutputDir.Text = dlg.SelectedPath;
        }

        private void btnExport_Click(object sender, EventArgs e)
        {
            // 验证输入
            if (string.IsNullOrEmpty(txtSpritePgk.Text) || !File.Exists(txtSpritePgk.Text))
            {
                MessageBox.Show("请选择有效的 sprite.pgk", "提示");
                return;
            }
            if (string.IsNullOrEmpty(txtSysPgk.Text) || !File.Exists(txtSysPgk.Text))
            {
                MessageBox.Show("请选择有效的 sys.pgk", "提示");
                return;
            }
            if (string.IsNullOrEmpty(txtOutputDir.Text) || !Directory.Exists(txtOutputDir.Text))
            {
                MessageBox.Show("请选择有效的输出目录", "提示");
                return;
            }

            // 解析装备ID
            int[] equipmentIds = new int[10];
            string[] parts = txtEquipment.Text.Split(',');
            if (parts.Length != 10)
            {
                MessageBox.Show("装备ID必须为10个数字，用逗号分隔", "提示");
                return;
            }
            for (int i = 0; i < 10; i++)
            {
                if (!int.TryParse(parts[i].Trim(), out equipmentIds[i]))
                {
                    MessageBox.Show($"第 {i + 1} 个装备ID无效：{parts[i]}", "提示");
                    return;
                }
            }

            // 开始导出
            btnExport.Enabled = false;
            progressBar.Value = 0;
            lblStatus.Text = "正在加载 PGK 包...";

            try
            {
                bool isMale = cmbGender.SelectedIndex == 0;
                int bodyId = (int)numBodyId.Value;
                int atdIndex = (int)numAtdIndex.Value;
                string charName = txtCharName.Text.Trim();

                // 加载 PGK 包
                var spritePgk = new PgkExtractor();
                spritePgk.Load(txtSpritePgk.Text);

                var sysPgk = new PgkExtractor();
                sysPgk.Load(txtSysPgk.Text);

                lblStatus.Text = "PGK 包加载完成，开始合成...";
                Application.DoEvents();

                // 创建转换器
                var converter = new AtzToPngConverter(spritePgk, sysPgk);

                // 输出目录结构：outputDir/characters/{charName}/...
                string outputBase = Path.Combine(txtOutputDir.Text, "characters");

                // 导出角色
                var manifest = converter.ExportCharacter(
                    isMale, bodyId, equipmentIds, atdIndex, outputBase, charName);

                progressBar.Value = 80;
                lblStatus.Text = "正在生成 manifest.json...";
                Application.DoEvents();

                // 生成 manifest.json
                var characters = new Dictionary<string, CharacterManifest>
                {
                    { charName, manifest }
                };

                string manifestPath = Path.Combine(outputBase, "manifest.json");
                AtzToPngConverter.GenerateManifest(manifestPath, characters);

                progressBar.Value = 100;
                lblStatus.Text = "导出完成！";
                lblProgress.Text = $"输出位置：{outputBase}\nmanifest.json：{manifestPath}";

                MessageBox.Show($"角色动画合成导出完成！\n\n输出目录：{outputBase}\n清单文件：{manifestPath}",
                    "导出成功", MessageBoxButtons.OK, MessageBoxIcon.Information);
            }
            catch (Exception ex)
            {
                lblStatus.Text = "导出失败";
                lblProgress.Text = ex.Message;
                MessageBox.Show($"导出失败：{ex.Message}\n\n{ex.StackTrace}",
                    "错误", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
            finally
            {
                btnExport.Enabled = true;
            }
        }
    }
}
