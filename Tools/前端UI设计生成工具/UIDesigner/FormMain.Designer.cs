using System.Drawing;
using System.Windows.Forms;

namespace UIDesigner
{
    partial class FormMain
    {
        private System.ComponentModel.IContainer components = null;
        private PictureBox picCanvas;
        private ComboBox ctlType, cboTemplate, cboStyle, folderSelect;
        private Button btnAddCtrl, btnApplyProp, btnPreviewCode, btnExport, btnClear;
        private Button btnSelectImg, btnClearImg, btnLoadTemplate;
        private Button btnBatchSelect, btnBatchClear, btnBatchStyle;
        private TextBox txtText, txtId, txtHtml, txtCss, txtJs, txtProjectRoot, txtFileName;
        private NumericUpDown numX, numY, numW, numH, numFont, numBorder, numRadius, numAlpha;
        private CheckBox chkHideDefault;
        private Label lblCtl, lblText, lblId, lblX, lblY, lblW, lblH;
        private Label lblFont, lblBorder, lblRadius, lblAlpha, lblStyle, lblTemplate;
        private Label lblRoot, lblFile;
        private Label lblGroup1, lblGroup2, lblGroup3, lblGroup4, lblGroup5;
        private Label lblHtml, lblCss, lblJs;

        protected override void Dispose(bool disposing)
        {
            if (disposing && components != null) components.Dispose();
            base.Dispose(disposing);
        }

        private void InitializeComponent()
        {
            picCanvas = new PictureBox();
            ctlType = new ComboBox();
            cboTemplate = new ComboBox();
            cboStyle = new ComboBox();
            folderSelect = new ComboBox();
            btnAddCtrl = new Button();
            btnApplyProp = new Button();
            btnPreviewCode = new Button();
            btnExport = new Button();
            btnClear = new Button();
            btnSelectImg = new Button();
            btnClearImg = new Button();
            btnLoadTemplate = new Button();
            btnBatchSelect = new Button();
            btnBatchClear = new Button();
            btnBatchStyle = new Button();
            txtText = new TextBox();
            txtId = new TextBox();
            txtHtml = new TextBox();
            txtCss = new TextBox();
            txtJs = new TextBox();
            txtProjectRoot = new TextBox();
            txtFileName = new TextBox();
            numX = new NumericUpDown();
            numY = new NumericUpDown();
            numW = new NumericUpDown();
            numH = new NumericUpDown();
            numFont = new NumericUpDown();
            numBorder = new NumericUpDown();
            numRadius = new NumericUpDown();
            numAlpha = new NumericUpDown();
            chkHideDefault = new CheckBox();
            lblCtl = new Label();
            lblText = new Label();
            lblId = new Label();
            lblX = new Label();
            lblY = new Label();
            lblW = new Label();
            lblH = new Label();
            lblFont = new Label();
            lblBorder = new Label();
            lblRadius = new Label();
            lblAlpha = new Label();
            lblStyle = new Label();
            lblTemplate = new Label();
            lblRoot = new Label();
            lblFile = new Label();
            lblGroup1 = new Label();
            lblGroup2 = new Label();
            lblGroup3 = new Label();
            lblGroup4 = new Label();
            lblGroup5 = new Label();
            lblHtml = new Label();
            lblCss = new Label();
            lblJs = new Label();
            ((System.ComponentModel.ISupportInitialize)picCanvas).BeginInit();
            ((System.ComponentModel.ISupportInitialize)numX).BeginInit();
            ((System.ComponentModel.ISupportInitialize)numY).BeginInit();
            ((System.ComponentModel.ISupportInitialize)numW).BeginInit();
            ((System.ComponentModel.ISupportInitialize)numH).BeginInit();
            ((System.ComponentModel.ISupportInitialize)numFont).BeginInit();
            ((System.ComponentModel.ISupportInitialize)numBorder).BeginInit();
            ((System.ComponentModel.ISupportInitialize)numRadius).BeginInit();
            ((System.ComponentModel.ISupportInitialize)numAlpha).BeginInit();
            SuspendLayout();
            // 
            // picCanvas
            // 
            picCanvas.BackColor = Color.FromArgb(35, 35, 35);
            picCanvas.BorderStyle = BorderStyle.FixedSingle;
            picCanvas.Location = new Point(10, 10);
            picCanvas.Name = "picCanvas";
            picCanvas.Size = new Size(800, 550);
            picCanvas.TabIndex = 0;
            picCanvas.TabStop = false;
            picCanvas.Paint += picCanvas_Paint;
            picCanvas.MouseDown += picCanvas_MouseDown;
            picCanvas.MouseMove += picCanvas_MouseMove;
            picCanvas.MouseUp += picCanvas_MouseUp;
            // 
            // ctlType
            // 
            ctlType.BackColor = Color.FromArgb(45, 45, 45);
            ctlType.DropDownStyle = ComboBoxStyle.DropDownList;
            ctlType.FlatStyle = FlatStyle.Flat;
            ctlType.ForeColor = Color.FromArgb(220, 220, 220);
            ctlType.FormattingEnabled = true;
            ctlType.Location = new Point(825, 64);
            ctlType.Name = "ctlType";
            ctlType.Size = new Size(280, 32);
            ctlType.TabIndex = 3;
            // 
            // cboTemplate
            // 
            cboTemplate.BackColor = Color.FromArgb(45, 45, 45);
            cboTemplate.DropDownStyle = ComboBoxStyle.DropDownList;
            cboTemplate.FlatStyle = FlatStyle.Flat;
            cboTemplate.ForeColor = Color.FromArgb(220, 220, 220);
            cboTemplate.FormattingEnabled = true;
            cboTemplate.Location = new Point(826, 438);
            cboTemplate.Name = "cboTemplate";
            cboTemplate.Size = new Size(175, 32);
            cboTemplate.TabIndex = 35;
            // 
            // cboStyle
            // 
            cboStyle.BackColor = Color.FromArgb(45, 45, 45);
            cboStyle.DropDownStyle = ComboBoxStyle.DropDownList;
            cboStyle.FlatStyle = FlatStyle.Flat;
            cboStyle.ForeColor = Color.FromArgb(220, 220, 220);
            cboStyle.FormattingEnabled = true;
            cboStyle.Location = new Point(826, 744);
            cboStyle.Name = "cboStyle";
            cboStyle.Size = new Size(280, 32);
            cboStyle.TabIndex = 32;
            cboStyle.SelectedIndexChanged += cboStyle_SelectedIndexChanged;
            // 
            // folderSelect
            // 
            folderSelect.Location = new Point(0, 0);
            folderSelect.Name = "folderSelect";
            folderSelect.Size = new Size(121, 32);
            folderSelect.TabIndex = 0;
            // 
            // btnAddCtrl
            // 
            btnAddCtrl.BackColor = Color.FromArgb(66, 133, 244);
            btnAddCtrl.FlatStyle = FlatStyle.Flat;
            btnAddCtrl.ForeColor = Color.White;
            btnAddCtrl.Location = new Point(825, 294);
            btnAddCtrl.Name = "btnAddCtrl";
            btnAddCtrl.Size = new Size(135, 32);
            btnAddCtrl.TabIndex = 16;
            btnAddCtrl.Text = "添加控件";
            btnAddCtrl.UseVisualStyleBackColor = false;
            btnAddCtrl.Click += btnAddCtrl_Click;
            // 
            // btnApplyProp
            // 
            btnApplyProp.BackColor = Color.FromArgb(45, 45, 45);
            btnApplyProp.FlatStyle = FlatStyle.Flat;
            btnApplyProp.ForeColor = Color.FromArgb(220, 220, 220);
            btnApplyProp.Location = new Point(970, 294);
            btnApplyProp.Name = "btnApplyProp";
            btnApplyProp.Size = new Size(135, 32);
            btnApplyProp.TabIndex = 17;
            btnApplyProp.Text = "应用属性";
            btnApplyProp.UseVisualStyleBackColor = false;
            btnApplyProp.Click += btnApplyProp_Click;
            // 
            // btnPreviewCode
            // 
            btnPreviewCode.BackColor = Color.FromArgb(66, 133, 244);
            btnPreviewCode.FlatStyle = FlatStyle.Flat;
            btnPreviewCode.ForeColor = Color.White;
            btnPreviewCode.Location = new Point(715, 571);
            btnPreviewCode.Name = "btnPreviewCode";
            btnPreviewCode.Size = new Size(95, 32);
            btnPreviewCode.TabIndex = 47;
            btnPreviewCode.Text = "生成代码";
            btnPreviewCode.UseVisualStyleBackColor = false;
            btnPreviewCode.Click += btnPreviewCode_Click;
            // 
            // btnExport
            // 
            btnExport.BackColor = Color.FromArgb(40, 167, 69);
            btnExport.FlatStyle = FlatStyle.Flat;
            btnExport.ForeColor = Color.White;
            btnExport.Location = new Point(1594, 811);
            btnExport.Name = "btnExport";
            btnExport.Size = new Size(120, 33);
            btnExport.TabIndex = 53;
            btnExport.Text = "一键导出";
            btnExport.UseVisualStyleBackColor = false;
            btnExport.Click += btnExport_Click;
            // 
            // btnClear
            // 
            btnClear.BackColor = Color.FromArgb(220, 53, 69);
            btnClear.FlatStyle = FlatStyle.Flat;
            btnClear.ForeColor = Color.White;
            btnClear.Location = new Point(825, 332);
            btnClear.Name = "btnClear";
            btnClear.Size = new Size(280, 32);
            btnClear.TabIndex = 18;
            btnClear.Text = "清空画布";
            btnClear.UseVisualStyleBackColor = false;
            btnClear.Click += btnClear_Click;
            // 
            // btnSelectImg
            // 
            btnSelectImg.BackColor = Color.FromArgb(45, 45, 45);
            btnSelectImg.FlatStyle = FlatStyle.Flat;
            btnSelectImg.ForeColor = Color.FromArgb(220, 220, 220);
            btnSelectImg.Location = new Point(826, 679);
            btnSelectImg.Name = "btnSelectImg";
            btnSelectImg.Size = new Size(135, 35);
            btnSelectImg.TabIndex = 29;
            btnSelectImg.Text = "选择贴图";
            btnSelectImg.UseVisualStyleBackColor = false;
            btnSelectImg.Click += btnSelectImg_Click;
            // 
            // btnClearImg
            // 
            btnClearImg.BackColor = Color.FromArgb(45, 45, 45);
            btnClearImg.FlatStyle = FlatStyle.Flat;
            btnClearImg.ForeColor = Color.FromArgb(220, 220, 220);
            btnClearImg.Location = new Point(971, 679);
            btnClearImg.Name = "btnClearImg";
            btnClearImg.Size = new Size(135, 35);
            btnClearImg.TabIndex = 30;
            btnClearImg.Text = "清空贴图";
            btnClearImg.UseVisualStyleBackColor = false;
            btnClearImg.Click += btnClearImg_Click;
            // 
            // btnLoadTemplate
            // 
            btnLoadTemplate.BackColor = Color.FromArgb(66, 133, 244);
            btnLoadTemplate.FlatStyle = FlatStyle.Flat;
            btnLoadTemplate.ForeColor = Color.White;
            btnLoadTemplate.Location = new Point(1006, 438);
            btnLoadTemplate.Name = "btnLoadTemplate";
            btnLoadTemplate.Size = new Size(100, 32);
            btnLoadTemplate.TabIndex = 36;
            btnLoadTemplate.Text = "加载";
            btnLoadTemplate.UseVisualStyleBackColor = false;
            btnLoadTemplate.Click += btnLoadTemplate_Click;
            // 
            // btnBatchSelect
            // 
            btnBatchSelect.BackColor = Color.FromArgb(45, 45, 45);
            btnBatchSelect.FlatStyle = FlatStyle.Flat;
            btnBatchSelect.ForeColor = Color.FromArgb(220, 220, 220);
            btnBatchSelect.Location = new Point(827, 476);
            btnBatchSelect.Name = "btnBatchSelect";
            btnBatchSelect.Size = new Size(90, 32);
            btnBatchSelect.TabIndex = 37;
            btnBatchSelect.Text = "全选";
            btnBatchSelect.UseVisualStyleBackColor = false;
            btnBatchSelect.Click += btnBatchSelect_Click;
            // 
            // btnBatchClear
            // 
            btnBatchClear.BackColor = Color.FromArgb(45, 45, 45);
            btnBatchClear.FlatStyle = FlatStyle.Flat;
            btnBatchClear.ForeColor = Color.FromArgb(220, 220, 220);
            btnBatchClear.Location = new Point(922, 476);
            btnBatchClear.Name = "btnBatchClear";
            btnBatchClear.Size = new Size(90, 32);
            btnBatchClear.TabIndex = 38;
            btnBatchClear.Text = "取消";
            btnBatchClear.UseVisualStyleBackColor = false;
            btnBatchClear.Click += btnBatchClear_Click;
            // 
            // btnBatchStyle
            // 
            btnBatchStyle.BackColor = Color.FromArgb(45, 45, 45);
            btnBatchStyle.FlatStyle = FlatStyle.Flat;
            btnBatchStyle.ForeColor = Color.FromArgb(220, 220, 220);
            btnBatchStyle.Location = new Point(1017, 476);
            btnBatchStyle.Name = "btnBatchStyle";
            btnBatchStyle.Size = new Size(90, 32);
            btnBatchStyle.TabIndex = 39;
            btnBatchStyle.Text = "批量样式";
            btnBatchStyle.UseVisualStyleBackColor = false;
            btnBatchStyle.Click += btnBatchStyle_Click;
            // 
            // txtText
            // 
            txtText.BackColor = Color.FromArgb(45, 45, 45);
            txtText.BorderStyle = BorderStyle.FixedSingle;
            txtText.ForeColor = Color.FromArgb(220, 220, 220);
            txtText.Location = new Point(825, 130);
            txtText.Name = "txtText";
            txtText.Size = new Size(280, 30);
            txtText.TabIndex = 5;
            // 
            // txtId
            // 
            txtId.BackColor = Color.FromArgb(45, 45, 45);
            txtId.BorderStyle = BorderStyle.FixedSingle;
            txtId.ForeColor = Color.FromArgb(220, 220, 220);
            txtId.Location = new Point(825, 190);
            txtId.Name = "txtId";
            txtId.Size = new Size(280, 30);
            txtId.TabIndex = 7;
            // 
            // txtHtml
            // 
            txtHtml.BackColor = Color.FromArgb(15, 15, 15);
            txtHtml.BorderStyle = BorderStyle.FixedSingle;
            txtHtml.Font = new Font("Consolas", 9F, FontStyle.Regular, GraphicsUnit.Point);
            txtHtml.ForeColor = Color.LightGreen;
            txtHtml.Location = new Point(12, 628);
            txtHtml.Multiline = true;
            txtHtml.Name = "txtHtml";
            txtHtml.ScrollBars = ScrollBars.Both;
            txtHtml.Size = new Size(400, 225);
            txtHtml.TabIndex = 42;
            // 
            // txtCss
            // 
            txtCss.BackColor = Color.FromArgb(15, 15, 15);
            txtCss.BorderStyle = BorderStyle.FixedSingle;
            txtCss.Font = new Font("Consolas", 9F, FontStyle.Regular, GraphicsUnit.Point);
            txtCss.ForeColor = Color.LightBlue;
            txtCss.Location = new Point(418, 628);
            txtCss.Multiline = true;
            txtCss.Name = "txtCss";
            txtCss.ScrollBars = ScrollBars.Both;
            txtCss.Size = new Size(400, 225);
            txtCss.TabIndex = 44;
            // 
            // txtJs
            // 
            txtJs.BackColor = Color.FromArgb(15, 15, 15);
            txtJs.BorderStyle = BorderStyle.FixedSingle;
            txtJs.Font = new Font("Consolas", 9F, FontStyle.Regular, GraphicsUnit.Point);
            txtJs.ForeColor = Color.LightYellow;
            txtJs.Location = new Point(1121, 37);
            txtJs.Multiline = true;
            txtJs.Name = "txtJs";
            txtJs.ScrollBars = ScrollBars.Both;
            txtJs.Size = new Size(593, 739);
            txtJs.TabIndex = 46;
            // 
            // txtProjectRoot
            // 
            txtProjectRoot.BackColor = Color.FromArgb(45, 45, 45);
            txtProjectRoot.BorderStyle = BorderStyle.FixedSingle;
            txtProjectRoot.ForeColor = Color.FromArgb(220, 220, 220);
            txtProjectRoot.Location = new Point(933, 813);
            txtProjectRoot.Name = "txtProjectRoot";
            txtProjectRoot.Size = new Size(345, 30);
            txtProjectRoot.TabIndex = 50;
            // 
            // txtFileName
            // 
            txtFileName.BackColor = Color.FromArgb(45, 45, 45);
            txtFileName.BorderStyle = BorderStyle.FixedSingle;
            txtFileName.ForeColor = Color.FromArgb(220, 220, 220);
            txtFileName.Location = new Point(1381, 813);
            txtFileName.Name = "txtFileName";
            txtFileName.Size = new Size(207, 30);
            txtFileName.TabIndex = 52;
            // 
            // numX
            // 
            numX.BackColor = Color.FromArgb(45, 45, 45);
            numX.BorderStyle = BorderStyle.FixedSingle;
            numX.ForeColor = Color.FromArgb(220, 220, 220);
            numX.Location = new Point(825, 258);
            numX.Maximum = new decimal(new int[] { 1000, 0, 0, 0 });
            numX.Name = "numX";
            numX.Size = new Size(60, 30);
            numX.TabIndex = 12;
            // 
            // numY
            // 
            numY.BackColor = Color.FromArgb(45, 45, 45);
            numY.BorderStyle = BorderStyle.FixedSingle;
            numY.ForeColor = Color.FromArgb(220, 220, 220);
            numY.Location = new Point(895, 258);
            numY.Maximum = new decimal(new int[] { 1000, 0, 0, 0 });
            numY.Name = "numY";
            numY.Size = new Size(60, 30);
            numY.TabIndex = 13;
            // 
            // numW
            // 
            numW.BackColor = Color.FromArgb(45, 45, 45);
            numW.BorderStyle = BorderStyle.FixedSingle;
            numW.ForeColor = Color.FromArgb(220, 220, 220);
            numW.Location = new Point(970, 258);
            numW.Maximum = new decimal(new int[] { 1000, 0, 0, 0 });
            numW.Minimum = new decimal(new int[] { 1, 0, 0, 0 });
            numW.Name = "numW";
            numW.Size = new Size(60, 30);
            numW.TabIndex = 14;
            numW.Value = new decimal(new int[] { 1, 0, 0, 0 });
            // 
            // numH
            // 
            numH.BackColor = Color.FromArgb(45, 45, 45);
            numH.BorderStyle = BorderStyle.FixedSingle;
            numH.ForeColor = Color.FromArgb(220, 220, 220);
            numH.Location = new Point(1045, 258);
            numH.Maximum = new decimal(new int[] { 1000, 0, 0, 0 });
            numH.Minimum = new decimal(new int[] { 1, 0, 0, 0 });
            numH.Name = "numH";
            numH.Size = new Size(60, 30);
            numH.TabIndex = 15;
            numH.Value = new decimal(new int[] { 1, 0, 0, 0 });
            // 
            // numFont
            // 
            numFont.BackColor = Color.FromArgb(45, 45, 45);
            numFont.BorderStyle = BorderStyle.FixedSingle;
            numFont.ForeColor = Color.FromArgb(220, 220, 220);
            numFont.Location = new Point(826, 581);
            numFont.Maximum = new decimal(new int[] { 120, 0, 0, 0 });
            numFont.Minimum = new decimal(new int[] { 1, 0, 0, 0 });
            numFont.Name = "numFont";
            numFont.Size = new Size(80, 30);
            numFont.TabIndex = 23;
            numFont.Value = new decimal(new int[] { 14, 0, 0, 0 });
            // 
            // numBorder
            // 
            numBorder.BackColor = Color.FromArgb(45, 45, 45);
            numBorder.BorderStyle = BorderStyle.FixedSingle;
            numBorder.ForeColor = Color.FromArgb(220, 220, 220);
            numBorder.Location = new Point(921, 581);
            numBorder.Maximum = new decimal(new int[] { 50, 0, 0, 0 });
            numBorder.Name = "numBorder";
            numBorder.Size = new Size(80, 30);
            numBorder.TabIndex = 24;
            // 
            // numRadius
            // 
            numRadius.BackColor = Color.FromArgb(45, 45, 45);
            numRadius.BorderStyle = BorderStyle.FixedSingle;
            numRadius.ForeColor = Color.FromArgb(220, 220, 220);
            numRadius.Location = new Point(1016, 581);
            numRadius.Maximum = new decimal(new int[] { 200, 0, 0, 0 });
            numRadius.Name = "numRadius";
            numRadius.Size = new Size(90, 30);
            numRadius.TabIndex = 25;
            // 
            // numAlpha
            // 
            numAlpha.BackColor = Color.FromArgb(45, 45, 45);
            numAlpha.BorderStyle = BorderStyle.FixedSingle;
            numAlpha.ForeColor = Color.FromArgb(220, 220, 220);
            numAlpha.Location = new Point(826, 643);
            numAlpha.Maximum = new decimal(new int[] { 255, 0, 0, 0 });
            numAlpha.Name = "numAlpha";
            numAlpha.Size = new Size(80, 30);
            numAlpha.TabIndex = 27;
            numAlpha.Value = new decimal(new int[] { 255, 0, 0, 0 });
            // 
            // chkHideDefault
            // 
            chkHideDefault.AutoSize = true;
            chkHideDefault.ForeColor = Color.FromArgb(220, 220, 220);
            chkHideDefault.Location = new Point(921, 645);
            chkHideDefault.Name = "chkHideDefault";
            chkHideDefault.Size = new Size(108, 28);
            chkHideDefault.TabIndex = 28;
            chkHideDefault.Text = "默认隐藏";
            chkHideDefault.UseVisualStyleBackColor = true;
            // 
            // lblCtl
            // 
            lblCtl.AutoSize = true;
            lblCtl.ForeColor = Color.FromArgb(150, 150, 150);
            lblCtl.Location = new Point(825, 37);
            lblCtl.Name = "lblCtl";
            lblCtl.Size = new Size(82, 24);
            lblCtl.TabIndex = 2;
            lblCtl.Text = "控件类型";
            // 
            // lblText
            // 
            lblText.AutoSize = true;
            lblText.ForeColor = Color.FromArgb(150, 150, 150);
            lblText.Location = new Point(825, 103);
            lblText.Name = "lblText";
            lblText.Size = new Size(82, 24);
            lblText.TabIndex = 4;
            lblText.Text = "显示文本";
            // 
            // lblId
            // 
            lblId.AutoSize = true;
            lblId.ForeColor = Color.FromArgb(150, 150, 150);
            lblId.Location = new Point(825, 163);
            lblId.Name = "lblId";
            lblId.Size = new Size(65, 24);
            lblId.TabIndex = 6;
            lblId.Text = "控件ID";
            // 
            // lblX
            // 
            lblX.AutoSize = true;
            lblX.ForeColor = Color.FromArgb(150, 150, 150);
            lblX.Location = new Point(825, 229);
            lblX.Name = "lblX";
            lblX.Size = new Size(22, 24);
            lblX.TabIndex = 8;
            lblX.Text = "X";
            // 
            // lblY
            // 
            lblY.AutoSize = true;
            lblY.ForeColor = Color.FromArgb(150, 150, 150);
            lblY.Location = new Point(895, 229);
            lblY.Name = "lblY";
            lblY.Size = new Size(21, 24);
            lblY.TabIndex = 9;
            lblY.Text = "Y";
            // 
            // lblW
            // 
            lblW.AutoSize = true;
            lblW.ForeColor = Color.FromArgb(150, 150, 150);
            lblW.Location = new Point(970, 229);
            lblW.Name = "lblW";
            lblW.Size = new Size(28, 24);
            lblW.TabIndex = 10;
            lblW.Text = "宽";
            // 
            // lblH
            // 
            lblH.AutoSize = true;
            lblH.ForeColor = Color.FromArgb(150, 150, 150);
            lblH.Location = new Point(1045, 229);
            lblH.Name = "lblH";
            lblH.Size = new Size(28, 24);
            lblH.TabIndex = 11;
            lblH.Text = "高";
            // 
            // lblFont
            // 
            lblFont.AutoSize = true;
            lblFont.ForeColor = Color.FromArgb(150, 150, 150);
            lblFont.Location = new Point(826, 556);
            lblFont.Name = "lblFont";
            lblFont.Size = new Size(46, 24);
            lblFont.TabIndex = 20;
            lblFont.Text = "字号";
            // 
            // lblBorder
            // 
            lblBorder.AutoSize = true;
            lblBorder.ForeColor = Color.FromArgb(150, 150, 150);
            lblBorder.Location = new Point(921, 556);
            lblBorder.Name = "lblBorder";
            lblBorder.Size = new Size(46, 24);
            lblBorder.TabIndex = 21;
            lblBorder.Text = "边框";
            // 
            // lblRadius
            // 
            lblRadius.AutoSize = true;
            lblRadius.ForeColor = Color.FromArgb(150, 150, 150);
            lblRadius.Location = new Point(1016, 556);
            lblRadius.Name = "lblRadius";
            lblRadius.Size = new Size(46, 24);
            lblRadius.TabIndex = 22;
            lblRadius.Text = "圆角";
            // 
            // lblAlpha
            // 
            lblAlpha.AutoSize = true;
            lblAlpha.ForeColor = Color.FromArgb(150, 150, 150);
            lblAlpha.Location = new Point(826, 616);
            lblAlpha.Name = "lblAlpha";
            lblAlpha.Size = new Size(64, 24);
            lblAlpha.TabIndex = 26;
            lblAlpha.Text = "透明度";
            // 
            // lblStyle
            // 
            lblStyle.AutoSize = true;
            lblStyle.ForeColor = Color.FromArgb(150, 150, 150);
            lblStyle.Location = new Point(826, 717);
            lblStyle.Name = "lblStyle";
            lblStyle.Size = new Size(82, 24);
            lblStyle.TabIndex = 31;
            lblStyle.Text = "风格预设";
            // 
            // lblTemplate
            // 
            lblTemplate.AutoSize = true;
            lblTemplate.ForeColor = Color.FromArgb(150, 150, 150);
            lblTemplate.Location = new Point(826, 411);
            lblTemplate.Name = "lblTemplate";
            lblTemplate.Size = new Size(64, 24);
            lblTemplate.TabIndex = 34;
            lblTemplate.Text = "UI模板";
            // 
            // lblRoot
            // 
            lblRoot.AutoSize = true;
            lblRoot.ForeColor = Color.FromArgb(150, 150, 150);
            lblRoot.Location = new Point(827, 819);
            lblRoot.Name = "lblRoot";
            lblRoot.Size = new Size(100, 24);
            lblRoot.TabIndex = 49;
            lblRoot.Text = "项目根目录";
            // 
            // lblFile
            // 
            lblFile.AutoSize = true;
            lblFile.ForeColor = Color.FromArgb(150, 150, 150);
            lblFile.Location = new Point(1284, 815);
            lblFile.Name = "lblFile";
            lblFile.Size = new Size(100, 24);
            lblFile.TabIndex = 51;
            lblFile.Text = "输出文件名";
            // 
            // lblGroup1
            // 
            lblGroup1.AutoSize = true;
            lblGroup1.Font = new Font("Microsoft Sans Serif", 8F, FontStyle.Bold, GraphicsUnit.Point);
            lblGroup1.ForeColor = Color.FromArgb(66, 133, 244);
            lblGroup1.Location = new Point(825, 9);
            lblGroup1.Name = "lblGroup1";
            lblGroup1.Size = new Size(111, 20);
            lblGroup1.TabIndex = 1;
            lblGroup1.Text = "【属性设置】";
            // 
            // lblGroup2
            // 
            lblGroup2.AutoSize = true;
            lblGroup2.Font = new Font("Microsoft Sans Serif", 8F, FontStyle.Bold, GraphicsUnit.Point);
            lblGroup2.ForeColor = Color.FromArgb(66, 133, 244);
            lblGroup2.Location = new Point(826, 528);
            lblGroup2.Name = "lblGroup2";
            lblGroup2.Size = new Size(111, 20);
            lblGroup2.TabIndex = 19;
            lblGroup2.Text = "【样式设置】";
            // 
            // lblGroup3
            // 
            lblGroup3.AutoSize = true;
            lblGroup3.Font = new Font("Microsoft Sans Serif", 8F, FontStyle.Bold, GraphicsUnit.Point);
            lblGroup3.ForeColor = Color.FromArgb(66, 133, 244);
            lblGroup3.Location = new Point(826, 383);
            lblGroup3.Name = "lblGroup3";
            lblGroup3.Size = new Size(128, 20);
            lblGroup3.TabIndex = 33;
            lblGroup3.Text = "【模板与批量】";
            // 
            // lblGroup4
            // 
            lblGroup4.AutoSize = true;
            lblGroup4.Font = new Font("Microsoft Sans Serif", 8F, FontStyle.Bold, GraphicsUnit.Point);
            lblGroup4.ForeColor = Color.FromArgb(66, 133, 244);
            lblGroup4.Location = new Point(12, 578);
            lblGroup4.Name = "lblGroup4";
            lblGroup4.Size = new Size(111, 20);
            lblGroup4.TabIndex = 40;
            lblGroup4.Text = "【代码预览】";
            // 
            // lblGroup5
            // 
            lblGroup5.AutoSize = true;
            lblGroup5.Font = new Font("Microsoft Sans Serif", 8F, FontStyle.Bold, GraphicsUnit.Point);
            lblGroup5.ForeColor = Color.FromArgb(66, 133, 244);
            lblGroup5.Location = new Point(827, 784);
            lblGroup5.Name = "lblGroup5";
            lblGroup5.Size = new Size(111, 20);
            lblGroup5.TabIndex = 48;
            lblGroup5.Text = "【导出设置】";
            // 
            // lblHtml
            // 
            lblHtml.AutoSize = true;
            lblHtml.ForeColor = Color.LightGreen;
            lblHtml.Location = new Point(12, 605);
            lblHtml.Name = "lblHtml";
            lblHtml.Size = new Size(61, 24);
            lblHtml.TabIndex = 41;
            lblHtml.Text = "HTML";
            // 
            // lblCss
            // 
            lblCss.AutoSize = true;
            lblCss.ForeColor = Color.LightBlue;
            lblCss.Location = new Point(418, 605);
            lblCss.Name = "lblCss";
            lblCss.Size = new Size(42, 24);
            lblCss.TabIndex = 43;
            lblCss.Text = "CSS";
            // 
            // lblJs
            // 
            lblJs.AutoSize = true;
            lblJs.ForeColor = Color.LightYellow;
            lblJs.Location = new Point(1121, 10);
            lblJs.Name = "lblJs";
            lblJs.Size = new Size(96, 24);
            lblJs.TabIndex = 45;
            lblJs.Text = "JavaScript";
            // 
            // FormMain
            // 
            BackColor = Color.FromArgb(20, 20, 20);
            ClientSize = new Size(1726, 865);
            Controls.Add(lblGroup5);
            Controls.Add(lblGroup4);
            Controls.Add(lblGroup3);
            Controls.Add(lblGroup2);
            Controls.Add(lblGroup1);
            Controls.Add(btnExport);
            Controls.Add(txtFileName);
            Controls.Add(lblFile);
            Controls.Add(txtProjectRoot);
            Controls.Add(lblRoot);
            Controls.Add(btnPreviewCode);
            Controls.Add(txtJs);
            Controls.Add(lblJs);
            Controls.Add(txtCss);
            Controls.Add(lblCss);
            Controls.Add(txtHtml);
            Controls.Add(lblHtml);
            Controls.Add(btnBatchStyle);
            Controls.Add(btnBatchClear);
            Controls.Add(btnBatchSelect);
            Controls.Add(btnLoadTemplate);
            Controls.Add(cboTemplate);
            Controls.Add(lblTemplate);
            Controls.Add(cboStyle);
            Controls.Add(lblStyle);
            Controls.Add(btnClearImg);
            Controls.Add(btnSelectImg);
            Controls.Add(chkHideDefault);
            Controls.Add(numAlpha);
            Controls.Add(lblAlpha);
            Controls.Add(numRadius);
            Controls.Add(lblRadius);
            Controls.Add(numBorder);
            Controls.Add(lblBorder);
            Controls.Add(numFont);
            Controls.Add(lblFont);
            Controls.Add(btnClear);
            Controls.Add(btnApplyProp);
            Controls.Add(btnAddCtrl);
            Controls.Add(numH);
            Controls.Add(lblH);
            Controls.Add(numW);
            Controls.Add(lblW);
            Controls.Add(numY);
            Controls.Add(lblY);
            Controls.Add(numX);
            Controls.Add(lblX);
            Controls.Add(txtId);
            Controls.Add(lblId);
            Controls.Add(txtText);
            Controls.Add(lblText);
            Controls.Add(ctlType);
            Controls.Add(lblCtl);
            Controls.Add(picCanvas);
            Name = "FormMain";
            StartPosition = FormStartPosition.CenterScreen;
            Text = "千年江湖 - UI设计生成工具";
            ((System.ComponentModel.ISupportInitialize)picCanvas).EndInit();
            ((System.ComponentModel.ISupportInitialize)numX).EndInit();
            ((System.ComponentModel.ISupportInitialize)numY).EndInit();
            ((System.ComponentModel.ISupportInitialize)numW).EndInit();
            ((System.ComponentModel.ISupportInitialize)numH).EndInit();
            ((System.ComponentModel.ISupportInitialize)numFont).EndInit();
            ((System.ComponentModel.ISupportInitialize)numBorder).EndInit();
            ((System.ComponentModel.ISupportInitialize)numRadius).EndInit();
            ((System.ComponentModel.ISupportInitialize)numAlpha).EndInit();
            ResumeLayout(false);
            PerformLayout();
        }
    }
}
