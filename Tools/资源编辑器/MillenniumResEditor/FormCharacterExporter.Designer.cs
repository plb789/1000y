namespace MillenniumResEditor
{
    partial class FormCharacterExporter
    {
        /// <summary>
        /// Required designer variable.
        /// </summary>
        private System.ComponentModel.IContainer components = null;

        /// <summary>
        /// Clean up any resources being used.
        /// </summary>
        /// <param name="disposing">true if managed resources should be disposed; otherwise, false.</param>
        protected override void Dispose(bool disposing)
        {
            if (disposing && (components != null))
            {
                components.Dispose();
            }
            base.Dispose(disposing);
        }

        #region Windows Form Designer generated code

        /// <summary>
        /// Required method for Designer support - do not modify
        /// the contents of this method with the code editor.
        /// </summary>
        private void InitializeComponent()
        {
            this.lblSpritePgk = new System.Windows.Forms.Label();
            this.lblSysPgk = new System.Windows.Forms.Label();
            this.lblOutputDir = new System.Windows.Forms.Label();
            this.txtSpritePgk = new System.Windows.Forms.TextBox();
            this.txtSysPgk = new System.Windows.Forms.TextBox();
            this.txtOutputDir = new System.Windows.Forms.TextBox();
            this.btnBrowseSprite = new System.Windows.Forms.Button();
            this.btnBrowseSys = new System.Windows.Forms.Button();
            this.btnBrowseOutput = new System.Windows.Forms.Button();
            this.btnExport = new System.Windows.Forms.Button();
            this.grpParams = new System.Windows.Forms.GroupBox();
            this.lblGender = new System.Windows.Forms.Label();
            this.cmbGender = new System.Windows.Forms.ComboBox();
            this.lblBodyId = new System.Windows.Forms.Label();
            this.numBodyId = new System.Windows.Forms.NumericUpDown();
            this.lblAtdIndex = new System.Windows.Forms.Label();
            this.numAtdIndex = new System.Windows.Forms.NumericUpDown();
            this.lblCharName = new System.Windows.Forms.Label();
            this.txtCharName = new System.Windows.Forms.TextBox();
            this.lblEquipment = new System.Windows.Forms.Label();
            this.txtEquipment = new System.Windows.Forms.TextBox();
            this.lblEquipHelp = new System.Windows.Forms.Label();
            this.progressBar = new System.Windows.Forms.ProgressBar();
            this.lblStatus = new System.Windows.Forms.Label();
            this.lblProgress = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.numBodyId)).BeginInit();
            ((System.ComponentModel.ISupportInitialize)(this.numAtdIndex)).BeginInit();
            this.grpParams.SuspendLayout();
            this.SuspendLayout();
            //
            // lblSpritePgk
            //
            this.lblSpritePgk.AutoSize = true;
            this.lblSpritePgk.Location = new System.Drawing.Point(20, 20);
            this.lblSpritePgk.Name = "lblSpritePgk";
            this.lblSpritePgk.Size = new System.Drawing.Size(120, 20);
            this.lblSpritePgk.TabIndex = 0;
            this.lblSpritePgk.Text = "sprite.pgk 路径:";
            //
            // lblSysPgk
            //
            this.lblSysPgk.AutoSize = true;
            this.lblSysPgk.Location = new System.Drawing.Point(20, 55);
            this.lblSysPgk.Name = "lblSysPgk";
            this.lblSysPgk.Size = new System.Drawing.Size(120, 20);
            this.lblSysPgk.TabIndex = 1;
            this.lblSysPgk.Text = "sys.pgk 路径:";
            //
            // lblOutputDir
            //
            this.lblOutputDir.AutoSize = true;
            this.lblOutputDir.Location = new System.Drawing.Point(20, 90);
            this.lblOutputDir.Name = "lblOutputDir";
            this.lblOutputDir.Size = new System.Drawing.Size(120, 20);
            this.lblOutputDir.TabIndex = 2;
            this.lblOutputDir.Text = "输出目录:";
            //
            // txtSpritePgk
            //
            this.txtSpritePgk.Location = new System.Drawing.Point(150, 18);
            this.txtSpritePgk.Name = "txtSpritePgk";
            this.txtSpritePgk.ReadOnly = true;
            this.txtSpritePgk.Size = new System.Drawing.Size(330, 24);
            this.txtSpritePgk.TabIndex = 3;
            //
            // txtSysPgk
            //
            this.txtSysPgk.Location = new System.Drawing.Point(150, 53);
            this.txtSysPgk.Name = "txtSysPgk";
            this.txtSysPgk.ReadOnly = true;
            this.txtSysPgk.Size = new System.Drawing.Size(330, 24);
            this.txtSysPgk.TabIndex = 4;
            //
            // txtOutputDir
            //
            this.txtOutputDir.Location = new System.Drawing.Point(150, 88);
            this.txtOutputDir.Name = "txtOutputDir";
            this.txtOutputDir.ReadOnly = true;
            this.txtOutputDir.Size = new System.Drawing.Size(330, 24);
            this.txtOutputDir.TabIndex = 5;
            //
            // btnBrowseSprite
            //
            this.btnBrowseSprite.Location = new System.Drawing.Point(490, 17);
            this.btnBrowseSprite.Name = "btnBrowseSprite";
            this.btnBrowseSprite.Size = new System.Drawing.Size(60, 26);
            this.btnBrowseSprite.TabIndex = 6;
            this.btnBrowseSprite.Text = "浏览...";
            this.btnBrowseSprite.UseVisualStyleBackColor = true;
            this.btnBrowseSprite.Click += new System.EventHandler(this.btnBrowseSprite_Click);
            //
            // btnBrowseSys
            //
            this.btnBrowseSys.Location = new System.Drawing.Point(490, 52);
            this.btnBrowseSys.Name = "btnBrowseSys";
            this.btnBrowseSys.Size = new System.Drawing.Size(60, 26);
            this.btnBrowseSys.TabIndex = 7;
            this.btnBrowseSys.Text = "浏览...";
            this.btnBrowseSys.UseVisualStyleBackColor = true;
            this.btnBrowseSys.Click += new System.EventHandler(this.btnBrowseSys_Click);
            //
            // btnBrowseOutput
            //
            this.btnBrowseOutput.Location = new System.Drawing.Point(490, 87);
            this.btnBrowseOutput.Name = "btnBrowseOutput";
            this.btnBrowseOutput.Size = new System.Drawing.Size(60, 26);
            this.btnBrowseOutput.TabIndex = 8;
            this.btnBrowseOutput.Text = "浏览...";
            this.btnBrowseOutput.UseVisualStyleBackColor = true;
            this.btnBrowseOutput.Click += new System.EventHandler(this.btnBrowseOutput_Click);
            //
            // btnExport
            //
            this.btnExport.BackColor = System.Drawing.Color.LightBlue;
            this.btnExport.Location = new System.Drawing.Point(20, 320);
            this.btnExport.Name = "btnExport";
            this.btnExport.Size = new System.Drawing.Size(530, 40);
            this.btnExport.TabIndex = 9;
            this.btnExport.Text = "开始合成导出";
            this.btnExport.UseVisualStyleBackColor = false;
            this.btnExport.Click += new System.EventHandler(this.btnExport_Click);
            //
            // grpParams
            //
            this.grpParams.Controls.Add(this.lblEquipHelp);
            this.grpParams.Controls.Add(this.txtEquipment);
            this.grpParams.Controls.Add(this.lblEquipment);
            this.grpParams.Controls.Add(this.txtCharName);
            this.grpParams.Controls.Add(this.lblCharName);
            this.grpParams.Controls.Add(this.numAtdIndex);
            this.grpParams.Controls.Add(this.lblAtdIndex);
            this.grpParams.Controls.Add(this.numBodyId);
            this.grpParams.Controls.Add(this.lblBodyId);
            this.grpParams.Controls.Add(this.cmbGender);
            this.grpParams.Controls.Add(this.lblGender);
            this.grpParams.Location = new System.Drawing.Point(20, 125);
            this.grpParams.Name = "grpParams";
            this.grpParams.Size = new System.Drawing.Size(530, 180);
            this.grpParams.TabIndex = 10;
            this.grpParams.TabStop = false;
            this.grpParams.Text = "导出参数";
            //
            // lblGender
            //
            this.lblGender.AutoSize = true;
            this.lblGender.Location = new System.Drawing.Point(20, 25);
            this.lblGender.Name = "lblGender";
            this.lblGender.Size = new System.Drawing.Size(60, 20);
            this.lblGender.TabIndex = 0;
            this.lblGender.Text = "性别:";
            //
            // cmbGender
            //
            this.cmbGender.DropDownStyle = System.Windows.Forms.ComboBoxStyle.DropDownList;
            this.cmbGender.FormattingEnabled = true;
            this.cmbGender.Items.AddRange(new object[] { "男性(n前缀)", "女性(a前缀)" });
            this.cmbGender.Location = new System.Drawing.Point(80, 22);
            this.cmbGender.Name = "cmbGender";
            this.cmbGender.Size = new System.Drawing.Size(100, 24);
            this.cmbGender.TabIndex = 1;
            //
            // lblBodyId
            //
            this.lblBodyId.AutoSize = true;
            this.lblBodyId.Location = new System.Drawing.Point(200, 25);
            this.lblBodyId.Name = "lblBodyId";
            this.lblBodyId.Size = new System.Drawing.Size(60, 20);
            this.lblBodyId.TabIndex = 2;
            this.lblBodyId.Text = "身体ID:";
            //
            // numBodyId
            //
            this.numBodyId.Location = new System.Drawing.Point(260, 22);
            this.numBodyId.Maximum = new decimal(new int[] { 99, 0, 0, 0 });
            this.numBodyId.Name = "numBodyId";
            this.numBodyId.Size = new System.Drawing.Size(60, 24);
            this.numBodyId.TabIndex = 3;
            this.numBodyId.Value = new decimal(new int[] { 1, 0, 0, 0 });
            //
            // lblAtdIndex
            //
            this.lblAtdIndex.AutoSize = true;
            this.lblAtdIndex.Location = new System.Drawing.Point(340, 25);
            this.lblAtdIndex.Name = "lblAtdIndex";
            this.lblAtdIndex.Size = new System.Drawing.Size(70, 20);
            this.lblAtdIndex.TabIndex = 4;
            this.lblAtdIndex.Text = "ATD索引:";
            //
            // numAtdIndex
            //
            this.numAtdIndex.Location = new System.Drawing.Point(410, 22);
            this.numAtdIndex.Maximum = new decimal(new int[] { 99, 0, 0, 0 });
            this.numAtdIndex.Name = "numAtdIndex";
            this.numAtdIndex.Size = new System.Drawing.Size(60, 24);
            this.numAtdIndex.TabIndex = 5;
            //
            // lblCharName
            //
            this.lblCharName.AutoSize = true;
            this.lblCharName.Location = new System.Drawing.Point(20, 60);
            this.lblCharName.Name = "lblCharName";
            this.lblCharName.Size = new System.Drawing.Size(60, 20);
            this.lblCharName.TabIndex = 6;
            this.lblCharName.Text = "导出名:";
            //
            // txtCharName
            //
            this.txtCharName.Location = new System.Drawing.Point(80, 57);
            this.txtCharName.Name = "txtCharName";
            this.txtCharName.Size = new System.Drawing.Size(100, 24);
            this.txtCharName.TabIndex = 7;
            this.txtCharName.Text = "male";
            //
            // lblEquipment
            //
            this.lblEquipment.AutoSize = true;
            this.lblEquipment.Location = new System.Drawing.Point(200, 60);
            this.lblEquipment.Name = "lblEquipment";
            this.lblEquipment.Size = new System.Drawing.Size(200, 20);
            this.lblEquipment.TabIndex = 8;
            this.lblEquipment.Text = "装备ID (0-9,逗号分隔):";
            //
            // txtEquipment
            //
            this.txtEquipment.Location = new System.Drawing.Point(200, 82);
            this.txtEquipment.Name = "txtEquipment";
            this.txtEquipment.Size = new System.Drawing.Size(310, 24);
            this.txtEquipment.TabIndex = 9;
            this.txtEquipment.Text = "0,0,0,0,0,0,0,0,0,0";
            //
            // lblEquipHelp
            //
            this.lblEquipHelp.ForeColor = System.Drawing.Color.Gray;
            this.lblEquipHelp.Location = new System.Drawing.Point(20, 115);
            this.lblEquipHelp.Name = "lblEquipHelp";
            this.lblEquipHelp.Size = new System.Drawing.Size(500, 40);
            this.lblEquipHelp.TabIndex = 10;
            this.lblEquipHelp.Text = "顺序: 0=身体 1-5基础装备 6=武器 7=头饰 8=头套 9=坐骑 (=不显示)";
            //
            // progressBar
            //
            this.progressBar.Location = new System.Drawing.Point(20, 375);
            this.progressBar.Maximum = 100;
            this.progressBar.Name = "progressBar";
            this.progressBar.Size = new System.Drawing.Size(530, 20);
            this.progressBar.TabIndex = 11;
            //
            // lblStatus
            //
            this.lblStatus.AutoSize = true;
            this.lblStatus.Location = new System.Drawing.Point(20, 405);
            this.lblStatus.Name = "lblStatus";
            this.lblStatus.Size = new System.Drawing.Size(530, 20);
            this.lblStatus.TabIndex = 12;
            this.lblStatus.Text = "就绪";
            //
            // lblProgress
            //
            this.lblProgress.ForeColor = System.Drawing.Color.Gray;
            this.lblProgress.Location = new System.Drawing.Point(20, 430);
            this.lblProgress.Name = "lblProgress";
            this.lblProgress.Size = new System.Drawing.Size(530, 40);
            this.lblProgress.TabIndex = 13;
            //
            // FormCharacterExporter
            //
            this.AutoScaleDimensions = new System.Drawing.SizeF(7F, 15F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(584, 481);
            this.Controls.Add(this.lblProgress);
            this.Controls.Add(this.lblStatus);
            this.Controls.Add(this.progressBar);
            this.Controls.Add(this.grpParams);
            this.Controls.Add(this.btnExport);
            this.Controls.Add(this.btnBrowseOutput);
            this.Controls.Add(this.txtOutputDir);
            this.Controls.Add(this.lblOutputDir);
            this.Controls.Add(this.btnBrowseSys);
            this.Controls.Add(this.txtSysPgk);
            this.Controls.Add(this.lblSysPgk);
            this.Controls.Add(this.btnBrowseSprite);
            this.Controls.Add(this.txtSpritePgk);
            this.Controls.Add(this.lblSpritePgk);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.MaximizeBox = false;
            this.Name = "FormCharacterExporter";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterParent;
            this.Text = "角色动画合成导出工具";
            ((System.ComponentModel.ISupportInitialize)(this.numBodyId)).EndInit();
            ((System.ComponentModel.ISupportInitialize)(this.numAtdIndex)).EndInit();
            this.grpParams.ResumeLayout(false);
            this.grpParams.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();
        }

        #endregion

        private System.Windows.Forms.Label lblSpritePgk;
        private System.Windows.Forms.Label lblSysPgk;
        private System.Windows.Forms.Label lblOutputDir;
        private System.Windows.Forms.TextBox txtSpritePgk;
        private System.Windows.Forms.TextBox txtSysPgk;
        private System.Windows.Forms.TextBox txtOutputDir;
        private System.Windows.Forms.Button btnBrowseSprite;
        private System.Windows.Forms.Button btnBrowseSys;
        private System.Windows.Forms.Button btnBrowseOutput;
        private System.Windows.Forms.Button btnExport;
        private System.Windows.Forms.GroupBox grpParams;
        private System.Windows.Forms.Label lblGender;
        private System.Windows.Forms.ComboBox cmbGender;
        private System.Windows.Forms.Label lblBodyId;
        private System.Windows.Forms.NumericUpDown numBodyId;
        private System.Windows.Forms.Label lblAtdIndex;
        private System.Windows.Forms.NumericUpDown numAtdIndex;
        private System.Windows.Forms.Label lblCharName;
        private System.Windows.Forms.TextBox txtCharName;
        private System.Windows.Forms.Label lblEquipment;
        private System.Windows.Forms.TextBox txtEquipment;
        private System.Windows.Forms.Label lblEquipHelp;
        private System.Windows.Forms.ProgressBar progressBar;
        private System.Windows.Forms.Label lblStatus;
        private System.Windows.Forms.Label lblProgress;
    }
}
