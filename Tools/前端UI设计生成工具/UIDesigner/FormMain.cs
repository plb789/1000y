using System;
using System.Collections.Generic;
using System.Drawing;
using System.IO;
using System.Linq;
using System.Windows.Forms;
using UIDesigner.Utils;

namespace UIDesigner
{
    public partial class FormMain : Form
    {
        private const int CanvasW = 800;
        private const int CanvasH = 600;
        private readonly List<ControlItem> _controlList = new List<ControlItem>();

        // 拖拽变量
        private bool _isDrag;
        private Point _dragStart;
        private ControlItem? _dragItem;

        public FormMain()
        {
            InitializeComponent();
            InitUI();
        }

        private void InitUI()
        {
            // 画布初始化
            picCanvas.Width = CanvasW;
            picCanvas.Height = CanvasH;
            picCanvas.BackColor = Color.FromArgb(40, 40, 40);

            // 控件类型
            ctlType.Items.AddRange(new object[]
            {
                "Label 文本",
                "Button 按钮",
                "TextBox 输入框",
                "Panel 面板",
                "ListBox 列表",
                "ImageBox 图片"
            });
            ctlType.SelectedIndex = 0;

            // 模板下拉
            cboTemplate.Items.AddRange(TemplateManager.GetTemplateNames().ToArray());
            // 导出目录
            folderSelect.Items.AddRange(new object[]
            {
                "Window 窗口","Widget 控件","ChatUI 聊天","RoleUI 角色","ShopUI 商店","SocialUI 社交"
            });
            folderSelect.SelectedIndex = 0;

            // 复古风格预设
            cboStyle.Items.Add("默认样式");
            cboStyle.Items.Add("千年复古棕");
            cboStyle.Items.Add("武侠红金");
            cboStyle.SelectedIndex = 0;

            txtProjectRoot.Text = @"D:\千年江湖_全栈网游\Frontend";
        }

        #region 画布绘制
        private void picCanvas_Paint(object sender, PaintEventArgs e)
        {
            Graphics g = e.Graphics;
            g.Clear(Color.FromArgb(40, 40, 40));
            foreach (var item in _controlList)
            {
                DrawHelper.DrawControl(g, item);
            }
        }
        #endregion

        #region 添加控件
        private void btnAddCtrl_Click(object sender, EventArgs e)
        {
            var type = GetCtrlType();
            var newItem = new ControlItem
            {
                CtrlType = type,
                CtrlId = $"ui_{Guid.NewGuid():N}",
                Text = txtText.Text,
                X = 50, Y = 50, Width = 120, Height = 35,
                BgColor = Color.White,
                FontColor = Color.Black,
                FontSize = 12,
                BorderWidth = 1,
                BorderColor = Color.Gray,
                Radius = 0,
                Opacity = 255
            };
            _controlList.Add(newItem);
            picCanvas.Invalidate();
        }

        private UiControlType GetCtrlType()
        {
            return ctlType.Text switch
            {
                "Label 文本" => UiControlType.Label,
                "Button 按钮" => UiControlType.Button,
                "TextBox 输入框" => UiControlType.TextBox,
                "Panel 面板" => UiControlType.Panel,
                "ListBox 列表" => UiControlType.ListBox,
                "ImageBox 图片" => UiControlType.ImageBox,
                _ => UiControlType.Label
            };
        }
        #endregion

        #region 鼠标拖拽选中
        private void picCanvas_MouseDown(object sender, MouseEventArgs e)
        {
            _controlList.ForEach(x => x.IsSelected = false);
            for (int i = _controlList.Count - 1; i >= 0; i--)
            {
                var item = _controlList[i];
                if (item.GetRect().Contains(e.Location))
                {
                    item.IsSelected = true;
                    _isDrag = true;
                    _dragStart = e.Location;
                    _dragItem = item;
                    SyncProperty(item);
                    break;
                }
            }
            picCanvas.Invalidate();
        }

        private void picCanvas_MouseMove(object sender, MouseEventArgs e)
        {
            if (!_isDrag || _dragItem == null) return;
            int dx = e.X - _dragStart.X;
            int dy = e.Y - _dragStart.Y;
            _dragItem.X += dx;
            _dragItem.Y += dy;
            _dragStart = e.Location;
            picCanvas.Invalidate();
        }

        private void picCanvas_MouseUp(object sender, MouseEventArgs e)
        {
            _isDrag = false;
            _dragItem = null;
        }
        #endregion

        #region 属性同步 & 样式应用
        private void SyncProperty(ControlItem item)
        {
            txtId.Text = item.CtrlId;
            txtText.Text = item.Text;
            numX.Value = item.X;
            numY.Value = item.Y;
            numW.Value = item.Width;
            numH.Value = item.Height;
            numFont.Value = item.FontSize;
            numBorder.Value = item.BorderWidth;
            numRadius.Value = item.Radius;
            numAlpha.Value = item.Opacity;
            chkHideDefault.Checked = item.IsHideDefault;
        }

        private void btnApplyProp_Click(object sender, EventArgs e)
        {
            var sel = _controlList.FirstOrDefault(x => x.IsSelected);
            if (sel == null) return;

            sel.CtrlId = txtId.Text;
            sel.Text = txtText.Text;
            sel.X = (int)numX.Value;
            sel.Y = (int)numY.Value;
            sel.Width = (int)numW.Value;
            sel.Height = (int)numH.Value;
            sel.FontSize = (int)numFont.Value;
            sel.BorderWidth = (int)numBorder.Value;
            sel.Radius = (int)numRadius.Value;
            sel.Opacity = (int)numAlpha.Value;
            sel.IsHideDefault = chkHideDefault.Checked;

            picCanvas.Invalidate();
        }

        // 选择背景贴图（DDS/PNG）
        private void btnSelectImg_Click(object sender, EventArgs e)
        {
            using (var dlg = new OpenFileDialog())
            {
                dlg.Filter = "图片文件|*.png;*.dds|所有文件|*.*";
                if (dlg.ShowDialog() != DialogResult.OK) return;
                var sel = _controlList.FirstOrDefault(x => x.IsSelected);
                if (sel == null) return;

                ImageLoader.DisposeBitmap(sel.BgImage);
                sel.ImagePath = dlg.FileName;
                sel.BgImage = ImageLoader.LoadImage(dlg.FileName);
                picCanvas.Invalidate();
            }
        }

        // 清除贴图
        private void btnClearImg_Click(object sender, EventArgs e)
        {
            var sel = _controlList.FirstOrDefault(x => x.IsSelected);
            if (sel == null) return;
            ImageLoader.DisposeBitmap(sel.BgImage);
            sel.ImagePath = "";
            sel.BgImage = null;
            picCanvas.Invalidate();
        }

        // 复古样式预设
        private void cboStyle_SelectedIndexChanged(object sender, EventArgs e)
        {
            var selList = _controlList.Where(x => x.IsSelected).ToList();
            if (selList.Count == 0) return;
            switch (cboStyle.Text)
            {
                case "千年复古棕":
                    selList.ForEach(x =>
                    {
                        x.BgColor = Color.FromArgb(70, 45, 20);
                        x.BorderColor = Color.Gold;
                        x.FontColor = Color.White;
                    });
                    break;
                case "武侠红金":
                    selList.ForEach(x =>
                    {
                        x.BgColor = Color.FromArgb(100, 20, 20);
                        x.BorderColor = Color.Gold;
                        x.FontColor = Color.Gold;
                    });
                    break;
            }
            picCanvas.Invalidate();
        }
        #endregion

        #region UI 模板库
        private void btnLoadTemplate_Click(object sender, EventArgs e)
        {
            string tempName = cboTemplate.Text;
            if (string.IsNullOrEmpty(tempName)) return;
            var newList = TemplateManager.LoadTemplate(tempName);
            _controlList.Clear();
            _controlList.AddRange(newList);
            picCanvas.Invalidate();
            MessageBox.Show($"模板【{tempName}】加载完成");
        }
        #endregion

        #region 批量编辑
        private void btnBatchSelect_Click(object sender, EventArgs e)
        {
            _controlList.ForEach(x => x.IsSelected = true);
            picCanvas.Invalidate();
        }

        private void btnBatchClear_Click(object sender, EventArgs e)
        {
            _controlList.ForEach(x => x.IsSelected = false);
            picCanvas.Invalidate();
        }

        private void btnBatchStyle_Click(object sender, EventArgs e)
        {
            var sel = _controlList.Where(x => x.IsSelected).ToList();
            if (sel.Count == 0)
            {
                MessageBox.Show("请先选中控件");
                return;
            }
            int r = (int)numRadius.Value;
            int b = (int)numBorder.Value;
            sel.ForEach(x =>
            {
                x.Radius = r;
                x.BorderWidth = b;
            });
            picCanvas.Invalidate();
            MessageBox.Show("批量样式应用完成");
        }
        #endregion

        #region 代码预览 & 导出
        private void btnPreviewCode_Click(object sender, EventArgs e)
        {
            if (_controlList.Count == 0)
            {
                MessageBox.Show("请先添加控件");
                return;
            }
            var gen = new CodeGenerator(_controlList);
            txtHtml.Text = gen.GenerateHtml("游戏UI");
            txtCss.Text = gen.GenerateCss();
            txtJs.Text = gen.GenerateJs();
        }

        private void btnExport_Click(object sender, EventArgs e)
        {
            string root = txtProjectRoot.Text.Trim();
            if (!Directory.Exists(root))
            {
                MessageBox.Show("项目根目录不存在");
                return;
            }
            string fileName = txtFileName.Text.Trim();
            if (string.IsNullOrEmpty(fileName))
            {
                MessageBox.Show("请输入文件名");
                return;
            }

            var folder = GetFolderType(folderSelect.Text);
            var gen = new CodeGenerator(_controlList);
            string html = gen.GenerateHtml(fileName);
            string css = gen.GenerateCss();
            string js = gen.GenerateJs();

            bool ok = ExportHelper.ExportAll(root, folder, fileName, html, css, js);
            MessageBox.Show(ok ? "导出成功" : "导出失败");
        }

        private ExportHelper.UiFolderType GetFolderType(string text)
        {
            return text switch
            {
                "Window 窗口" => ExportHelper.UiFolderType.Window,
                "Widget 控件" => ExportHelper.UiFolderType.Widget,
                "ChatUI 聊天" => ExportHelper.UiFolderType.ChatUI,
                "RoleUI 角色" => ExportHelper.UiFolderType.RoleUI,
                "ShopUI 商店" => ExportHelper.UiFolderType.ShopUI,
                "SocialUI 社交" => ExportHelper.UiFolderType.SocialUI,
                _ => ExportHelper.UiFolderType.Window
            };
        }
        #endregion

        private void btnClear_Click(object sender, EventArgs e)
        {
            _controlList.Clear();
            txtHtml.Clear();
            txtCss.Clear();
            txtJs.Clear();
            picCanvas.Invalidate();
        }
    }
}