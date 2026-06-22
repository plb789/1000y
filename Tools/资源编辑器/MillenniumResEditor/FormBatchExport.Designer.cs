namespace MillenniumResEditor
{
    partial class FormBatchExport
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
            this.grpSource = new System.Windows.Forms.GroupBox();
            this.cmbMode = new System.Windows.Forms.ComboBox();
            this.lblMode = new System.Windows.Forms.Label();
            this.btnBrowseSysPgk = new System.Windows.Forms.Button();
            this.txtSysPgk = new System.Windows.Forms.TextBox();
            this.lblPgk2 = new System.Windows.Forms.Label();
            this.btnBrowsePgk = new System.Windows.Forms.Button();
            this.txtSpritePgk = new System.Windows.Forms.TextBox();
            this.lblPgk1 = new System.Windows.Forms.Label();
            this.btnBrowseAtdDir = new System.Windows.Forms.Button();
            this.txtAtdDir = new System.Windows.Forms.TextBox();
            this.lblAtdDir = new System.Windows.Forms.Label();
            this.btnBrowseAtzDir = new System.Windows.Forms.Button();
            this.txtAtzDir = new System.Windows.Forms.TextBox();
            this.lblAtzDir = new System.Windows.Forms.Label();
            this.grpParams = new System.Windows.Forms.GroupBox();
            this.chkDefaultEquip = new System.Windows.Forms.CheckBox();
            this.numMaxAtdIndex = new System.Windows.Forms.NumericUpDown();
            this.lblMaxAtd = new System.Windows.Forms.Label();
            this.numMaxBodyId = new System.Windows.Forms.NumericUpDown();
            this.lblMaxBody = new System.Windows.Forms.Label();
            this.grpOutput = new System.Windows.Forms.GroupBox();
            this.btnBrowseOutput = new System.Windows.Forms.Button();
            this.txtOutputDir = new System.Windows.Forms.TextBox();
            this.lblOutput = new System.Windows.Forms.Label();
            this.btnStartExport = new System.Windows.Forms.Button();
            this.lblProgress = new System.Windows.Forms.Label();
            this.prgProgress = new System.Windows.Forms.ProgressBar();
            this.rtbLog = new System.Windows.Forms.RichTextBox();
            this.lblStats = new System.Windows.Forms.Label();
            this.lblLogTitle = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.numMaxAtdIndex)).BeginInit();
            ((System.ComponentModel.ISupportInitialize)(this.numMaxBodyId)).BeginInit();
            this.grpSource.SuspendLayout();
            this.grpParams.SuspendLayout();
            this.grpOutput.SuspendLayout();
            this.SuspendLayout();
            //
            // grpSource
            //
            this.grpSource.Controls.Add(this.lblAtdDir);
            this.grpSource.Controls.Add(this.btnBrowseAtdDir);
            this.grpSource.Controls.Add(this.txtAtdDir);
            this.grpSource.Controls.Add(this.lblAtzDir);
            this.grpSource.Controls.Add(this.btnBrowseAtzDir);
            this.grpSource.Controls.Add(this.txtAtzDir);
            this.grpSource.Controls.Add(this.cmbMode);
            this.grpSource.Controls.Add(this.lblMode);
            this.grpSource.Controls.Add(this.btnBrowseSysPgk);
            this.grpSource.Controls.Add(this.txtSysPgk);
            this.grpSource.Controls.Add(this.lblPgk2);
            this.grpSource.Controls.Add(this.btnBrowsePgk);
            this.grpSource.Controls.Add(this.txtSpritePgk);
            this.grpSource.Controls.Add(this.lblPgk1);
            this.grpSource.Location = new System.Drawing.Point(10, 10);
            this.grpSource.Name = "grpSource";
            this.grpSource.Size = new System.Drawing.Size(640, 130);
            this.grpSource.TabIndex = 0;
            this.grpSource.TabStop = false;
            this.grpSource.Text = "数据源";
            //
            // cmbMode
            //
            this.cmbMode.DropDownStyle = System.Windows.Forms.ComboBoxStyle.DropDownList;
            this.cmbMode.FormattingEnabled = true;
            this.cmbMode.Items.AddRange(new object[] { "PGK 包模式", "独立文件目录模式" });
            this.cmbMode.Location = new System.Drawing.Point(60, 22);
            this.cmbMode.Name = "cmbMode";
            this.cmbMode.Size = new System.Drawing.Size(150, 24);
            this.cmbMode.TabIndex = 0;
            this.cmbMode.SelectedIndexChanged += new System.EventHandler(this.CmbMode_SelectedIndexChanged);
            //
            // lblMode
            //
            this.lblMode.AutoSize = true;
            this.lblMode.Location = new System.Drawing.Point(15, 25);
            this.lblMode.Name = "lblMode";
            this.lblMode.Size = new System.Drawing.Size(40, 20);
            this.lblMode.TabIndex = 1;
            this.lblMode.Text = "模式:";
            //
            // btnBrowseSysPgk
            //
            this.btnBrowseSysPgk.Location = new System.Drawing.Point(450, 78);
            this.btnBrowseSysPgk.Name = "btnBrowseSysPgk";
            this.btnBrowseSysPgk.Size = new System.Drawing.Size(35, 26);
            this.btnBrowseSysPgk.TabIndex = 14;
            this.btnBrowseSysPgk.Text = "...";
            this.btnBrowseSysPgk.UseVisualStyleBackColor = true;
            this.btnBrowseSysPgk.Click += new System.EventHandler(this.BtnBrowseSysPgk_Click);
            //
            // txtSysPgk
            //
            this.txtSysPgk.Location = new System.Drawing.Point(95, 80);
            this.txtSysPgk.Name = "txtSysPgk";
            this.txtSysPgk.ReadOnly = true;
            this.txtSysPgk.Size = new System.Drawing.Size(350, 22);
            this.txtSysPgk.TabIndex = 13;
            //
            // lblPgk2
            //
            this.lblPgk2.AutoSize = true;
            this.lblPgk2.Location = new System.Drawing.Point(15, 82);
            this.lblPgk2.Name = "lblPgk2";
            this.lblPgk2.Size = new System.Drawing.Size(80, 20);
            this.lblPgk2.TabIndex = 12;
            this.lblPgk2.Text = "sys.pgk:";
            //
            // btnBrowsePgk
            //
            this.btnBrowsePgk.Location = new System.Drawing.Point(450, 48);
            this.btnBrowsePgk.Name = "btnBrowsePgk";
            this.btnBrowsePgk.Size = new System.Drawing.Size(35, 26);
            this.btnBrowsePgk.TabIndex = 11;
            this.btnBrowsePgk.Text = "...";
            this.btnBrowsePgk.UseVisualStyleBackColor = true;
            this.btnBrowsePgk.Click += new System.EventHandler(this.BtnBrowsePgk_Click);
            //
            // txtSpritePgk
            //
            this.txtSpritePgk.Location = new System.Drawing.Point(95, 50);
            this.txtSpritePgk.Name = "txtSpritePgk";
            this.txtSpritePgk.ReadOnly = true;
            this.txtSpritePgk.Size = new System.Drawing.Size(350, 22);
            this.txtSpritePgk.TabIndex = 10;
            //
            // lblPgk1
            //
            this.lblPgk1.AutoSize = true;
            this.lblPgk1.Location = new System.Drawing.Point(15, 52);
            this.lblPgk1.Name = "lblPgk1";
            this.lblPgk1.Size = new System.Drawing.Size(80, 20);
            this.lblPgk1.TabIndex = 9;
            this.lblPgk1.Text = "sprite.pgk:";
            //
            // btnBrowseAtdDir
            //
            this.btnBrowseAtdDir.Location = new System.Drawing.Point(450, 78);
            this.btnBrowseAtdDir.Name = "btnBrowseAtdDir";
            this.btnBrowseAtdDir.Size = new System.Drawing.Size(35, 26);
            this.btnBrowseAtdDir.TabIndex = 19;
            this.btnBrowseAtdDir.Text = "...";
            this.btnBrowseAtdDir.UseVisualStyleBackColor = true;
            this.btnBrowseAtdDir.Visible = false;
            this.btnBrowseAtdDir.Click += new System.EventHandler(this.BtnBrowseAtdDir_Click);
            //
            // txtAtdDir
            //
            this.txtAtdDir.Location = new System.Drawing.Point(80, 80);
            this.txtAtdDir.Name = "txtAtdDir";
            this.txtAtdDir.Size = new System.Drawing.Size(365, 22);
            this.txtAtdDir.TabIndex = 18;
            this.txtAtdDir.Visible = false;
            //
            // lblAtdDir
            //
            this.lblAtdDir.AutoSize = true;
            this.lblAtdDir.Location = new System.Drawing.Point(15, 82);
            this.lblAtdDir.Name = "lblAtdDir";
            this.lblAtdDir.Size = new System.Drawing.Size(60, 20);
            this.lblAtdDir.TabIndex = 17;
            this.lblAtdDir.Text = "ATD 目录:";
            this.lblAtdDir.Visible = false;
            //
            // btnBrowseAtzDir
            //
            this.btnBrowseAtzDir.Location = new System.Drawing.Point(450, 48);
            this.btnBrowseAtzDir.Name = "btnBrowseAtzDir";
            this.btnBrowseAtzDir.Size = new System.Drawing.Size(35, 26);
            this.btnBrowseAtzDir.TabIndex = 16;
            this.btnBrowseAtzDir.Text = "...";
            this.btnBrowseAtzDir.UseVisualStyleBackColor = true;
            this.btnBrowseAtzDir.Visible = false;
            this.btnBrowseAtzDir.Click += new System.EventHandler(this.BtnBrowseAtzDir_Click);
            //
            // txtAtzDir
            //
            this.txtAtzDir.Location = new System.Drawing.Point(80, 50);
            this.txtAtzDir.Name = "txtAtzDir";
            this.txtAtzDir.Size = new System.Drawing.Size(365, 22);
            this.txtAtzDir.TabIndex = 15;
            this.txtAtzDir.Visible = false;
            //
            // lblAtzDir
            //
            this.lblAtzDir.AutoSize = true;
            this.lblAtzDir.Location = new System.Drawing.Point(15, 52);
            this.lblAtzDir.Name = "lblAtzDir";
            this.lblAtzDir.Size = new System.Drawing.Size(60, 20);
            this.lblAtzDir.TabIndex = 14;
            this.lblAtzDir.Text = "ATZ 目录:";
            this.lblAtzDir.Visible = false;
            //
            // grpParams
            //
            this.grpParams.Controls.Add(this.chkDefaultEquip);
            this.grpParams.Controls.Add(this.numMaxAtdIndex);
            this.grpParams.Controls.Add(this.lblMaxAtd);
            this.grpParams.Controls.Add(this.numMaxBodyId);
            this.grpParams.Controls.Add(this.lblMaxBody);
            this.grpParams.Location = new System.Drawing.Point(10, 145);
            this.grpParams.Name = "grpParams";
            this.grpParams.Size = new System.Drawing.Size(640, 85);
            this.grpParams.TabIndex = 1;
            this.grpParams.TabStop = false;
            this.grpParams.Text = "导出参数";
            //
            // chkDefaultEquip
            //
            this.chkDefaultEquip.AutoSize = true;
            this.chkDefaultEquip.Checked = true;
            this.chkDefaultEquip.CheckState = System.Windows.Forms.CheckState.Checked;
            this.chkDefaultEquip.Location = new System.Drawing.Point(375, 27);
            this.chkDefaultEquip.Name = "chkDefaultEquip";
            this.chkDefaultEquip.Size = new System.Drawing.Size(250, 20);
            this.chkDefaultEquip.TabIndex = 4;
            this.chkDefaultEquip.Text = "仅导出默认装备组合（裸体）";
            this.chkDefaultEquip.UseVisualStyleBackColor = true;
            //
            // numMaxAtdIndex
            //
            this.numMaxAtdIndex.Location = new System.Drawing.Point(290, 25);
            this.numMaxAtdIndex.Maximum = new decimal(new int[] { 99, 0, 0, 0 });
            this.numMaxAtdIndex.Name = "numMaxAtdIndex";
            this.numMaxAtdIndex.Size = new System.Drawing.Size(70, 22);
            this.numMaxAtdIndex.TabIndex = 3;
            this.numMaxAtdIndex.Value = new decimal(new int[] { 30, 0, 0, 0 });
            //
            // lblMaxAtd
            //
            this.lblMaxAtd.AutoSize = true;
            this.lblMaxAtd.Location = new System.Drawing.Point(195, 28);
            this.lblMaxAtd.Name = "lblMaxAtd";
            this.lblMaxAtd.Size = new System.Drawing.Size(90, 20);
            this.lblMaxAtd.TabIndex = 2;
            this.lblMaxAtd.Text = "最大ATD索引:";
            //
            // numMaxBodyId
            //
            this.numMaxBodyId.Location = new System.Drawing.Point(110, 25);
            this.numMaxBodyId.Maximum = new decimal(new int[] { 999, 0, 0, 0 });
            this.numMaxBodyId.Name = "numMaxBodyId";
            this.numMaxBodyId.Size = new System.Drawing.Size(70, 22);
            this.numMaxBodyId.TabIndex = 1;
            this.numMaxBodyId.Value = new decimal(new int[] { 200, 0, 0, 0 });
            //
            // lblMaxBody
            //
            this.lblMaxBody.AutoSize = true;
            this.lblMaxBody.Location = new System.Drawing.Point(15, 28);
            this.lblMaxBody.Name = "lblMaxBody";
            this.lblMaxBody.Size = new System.Drawing.Size(90, 20);
            this.lblMaxBody.TabIndex = 0;
            this.lblMaxBody.Text = "最大身体ID:";
            //
            // grpOutput
            //
            this.grpOutput.Controls.Add(this.btnBrowseOutput);
            this.grpOutput.Controls.Add(this.txtOutputDir);
            this.grpOutput.Controls.Add(this.lblOutput);
            this.grpOutput.Location = new System.Drawing.Point(10, 235);
            this.grpOutput.Name = "grpOutput";
            this.grpOutput.Size = new System.Drawing.Size(640, 55);
            this.grpOutput.TabIndex = 2;
            this.grpOutput.TabStop = false;
            this.grpOutput.Text = "输出设置";
            //
            // btnBrowseOutput
            //
            this.btnBrowseOutput.Location = new System.Drawing.Point(518, 21);
            this.btnBrowseOutput.Name = "btnBrowseOutput";
            this.btnBrowseOutput.Size = new System.Drawing.Size(70, 26);
            this.btnBrowseOutput.TabIndex = 2;
            this.btnBrowseOutput.Text = "浏览...";
            this.btnBrowseOutput.UseVisualStyleBackColor = true;
            this.btnBrowseOutput.Click += new System.EventHandler(this.BtnBrowseOutput_Click);
            //
            // txtOutputDir
            //
            this.txtOutputDir.Location = new System.Drawing.Point(82, 23);
            this.txtOutputDir.Name = "txtOutputDir";
            this.txtOutputDir.Size = new System.Drawing.Size(430, 22);
            this.txtOutputDir.TabIndex = 1;
            //
            // lblOutput
            //
            this.lblOutput.AutoSize = true;
            this.lblOutput.Location = new System.Drawing.Point(15, 25);
            this.lblOutput.Name = "lblOutput";
            this.lblOutput.Size = new System.Drawing.Size(65, 20);
            this.lblOutput.TabIndex = 0;
            this.lblOutput.Text = "输出目录:";
            //
            // btnStartExport
            //
            this.btnStartExport.BackColor = System.Drawing.Color.LightGreen;
            this.btnStartExport.Font = new System.Drawing.Font("Microsoft YaHei UI", 10F, System.Drawing.FontStyle.Bold);
            this.btnStartExport.Location = new System.Drawing.Point(10, 295);
            this.btnStartExport.Name = "btnStartExport";
            this.btnStartExport.Size = new System.Drawing.Size(160, 36);
            this.btnStartExport.TabIndex = 3;
            this.btnStartExport.Text = "▶ 开始批量导出";
            this.btnStartExport.UseVisualStyleBackColor = false;
            this.btnStartExport.Click += new System.EventHandler(this.BtnStartExport_Click);
            //
            // lblProgress
            //
            this.lblProgress.AutoSize = true;
            this.lblProgress.Location = new System.Drawing.Point(180, 305);
            this.lblProgress.Name = "lblProgress";
            this.lblProgress.Size = new System.Drawing.Size(400, 18);
            this.lblProgress.TabIndex = 4;
            this.lblProgress.Text = "就绪";
            //
            // prgProgress
            //
            this.prgProgress.Location = new System.Drawing.Point(180, 328);
            this.prgProgress.Name = "prgProgress";
            this.prgProgress.Size = new System.Drawing.Size(470, 22);
            this.prgProgress.Style = System.Windows.Forms.ProgressBarStyle.Continuous;
            this.prgProgress.TabIndex = 5;
            //
            // rtbLog
            //
            this.rtbLog.Font = new System.Drawing.Font("Consolas", 9F);
            this.rtbLog.Location = new System.Drawing.Point(10, 360);
            this.rtbLog.Name = "rtbLog";
            this.rtbLog.ReadOnly = true;
            this.rtbLog.Size = new System.Drawing.Size(640, 140);
            this.rtbLog.TabIndex = 7;
            this.rtbLog.Text = "";
            //
            // lblStats
            //
            this.lblStats.ForeColor = System.Drawing.Color.DarkBlue;
            this.lblStats.Location = new System.Drawing.Point(10, 510);
            this.lblStats.Name = "lblStats";
            this.lblStats.Size = new System.Drawing.Size(640, 20);
            this.lblStats.TabIndex = 8;
            this.lblStats.Text = "";
            //
            // lblLogTitle
            //
            this.lblLogTitle.AutoSize = true;
            this.lblLogTitle.Location = new System.Drawing.Point(10, 340);
            this.lblLogTitle.Name = "lblLogTitle";
            this.lblLogTitle.Size = new System.Drawing.Size(70, 18);
            this.lblLogTitle.TabIndex = 6;
            this.lblLogTitle.Text = "导出日志:";
            //
            // FormBatchExport
            //
            this.AutoScaleDimensions = new System.Drawing.SizeF(7F, 15F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(680, 580);
            this.Controls.Add(this.lblStats);
            this.Controls.Add(this.rtbLog);
            this.Controls.Add(this.lblLogTitle);
            this.Controls.Add(this.prgProgress);
            this.Controls.Add(this.lblProgress);
            this.Controls.Add(this.btnStartExport);
            this.Controls.Add(this.grpOutput);
            this.Controls.Add(this.grpParams);
            this.Controls.Add(this.grpSource);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.MaximizeBox = false;
            this.Name = "FormBatchExport";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterParent;
            this.Text = "批量角色库导出工具";
            ((System.ComponentModel.ISupportInitialize)(this.numMaxAtdIndex)).EndInit();
            ((System.ComponentModel.ISupportInitialize)(this.numMaxBodyId)).EndInit();
            this.grpSource.ResumeLayout(false);
            this.grpSource.PerformLayout();
            this.grpParams.ResumeLayout(false);
            this.grpParams.PerformLayout();
            this.grpOutput.ResumeLayout(false);
            this.grpOutput.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();
        }

        #endregion

        private System.Windows.Forms.ComboBox cmbMode;
        private System.Windows.Forms.TextBox txtSpritePgk;
        private System.Windows.Forms.Button btnBrowsePgk;
        private System.Windows.Forms.TextBox txtSysPgk;
        private System.Windows.Forms.Button btnBrowseSysPgk;
        private System.Windows.Forms.TextBox txtAtzDir;
        private System.Windows.Forms.Button btnBrowseAtzDir;
        private System.Windows.Forms.TextBox txtAtdDir;
        private System.Windows.Forms.Button btnBrowseAtdDir;
        private System.Windows.Forms.NumericUpDown numMaxBodyId;
        private System.Windows.Forms.NumericUpDown numMaxAtdIndex;
        private System.Windows.Forms.CheckBox chkDefaultEquip;
        private System.Windows.Forms.TextBox txtOutputDir;
        private System.Windows.Forms.Button btnBrowseOutput;
        private System.Windows.Forms.Button btnStartExport;
        private System.Windows.Forms.ProgressBar prgProgress;
        private System.Windows.Forms.Label lblProgress;
        private System.Windows.Forms.RichTextBox rtbLog;
        private System.Windows.Forms.Label lblStats;
        private System.Windows.Forms.GroupBox grpSource;
        private System.Windows.Forms.GroupBox grpParams;
        private System.Windows.Forms.GroupBox grpOutput;
        private System.Windows.Forms.Label lblMode;
        private System.Windows.Forms.Label lblPgk1;
        private System.Windows.Forms.Label lblPgk2;
        private System.Windows.Forms.Label lblAtzDir;
        private System.Windows.Forms.Label lblAtdDir;
        private System.Windows.Forms.Label lblMaxBody;
        private System.Windows.Forms.Label lblMaxAtd;
        private System.Windows.Forms.Label lblOutput;
        private System.Windows.Forms.Label lblLogTitle;
    }
}
