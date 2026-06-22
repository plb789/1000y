namespace MillenniumResEditor
{
    partial class FormAtzViewer
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
            this.btnOpenAtz = new System.Windows.Forms.Button();
            this.btnOpenPgk = new System.Windows.Forms.Button();
            this.btnExportPng = new System.Windows.Forms.Button();
            this.btnExportCharacter = new System.Windows.Forms.Button();
            this.lblInfo = new System.Windows.Forms.Label();
            this.picPreview = new System.Windows.Forms.PictureBox();
            this.lstFrames = new System.Windows.Forms.ListBox();
            this.lstPgkFiles = new System.Windows.Forms.ListBox();
            this.trkFrame = new System.Windows.Forms.TrackBar();
            this.grpTools = new System.Windows.Forms.GroupBox();
            this.btnPlayAnimation = new System.Windows.Forms.Button();
            this.btnStopAnimation = new System.Windows.Forms.Button();
            this.btnExportGif = new System.Windows.Forms.Button();
            this.lblFps = new System.Windows.Forms.Label();
            this.numFps = new System.Windows.Forms.NumericUpDown();
            this.chkLoop = new System.Windows.Forms.CheckBox();
            this.btnExportCurrentFrame = new System.Windows.Forms.Button();
            this.btnBatchExport = new System.Windows.Forms.Button();
            ((System.ComponentModel.ISupportInitialize)(this.picPreview)).BeginInit();
            ((System.ComponentModel.ISupportInitialize)(this.trkFrame)).BeginInit();
            ((System.ComponentModel.ISupportInitialize)(this.numFps)).BeginInit();
            this.grpTools.SuspendLayout();
            this.SuspendLayout();
            //
            // 第一行按钮（左侧功能按钮）
            //
            this.btnOpenAtz.Location = new System.Drawing.Point(10, 12);
            this.btnOpenAtz.Name = "btnOpenAtz";
            this.btnOpenAtz.Size = new System.Drawing.Size(90, 30);
            this.btnOpenAtz.TabIndex = 0;
            this.btnOpenAtz.Text = "打开 .atz";
            this.btnOpenAtz.UseVisualStyleBackColor = true;
            this.btnOpenAtz.Click += new System.EventHandler(this.btnOpenAtz_Click);
            //
            this.btnOpenPgk.Location = new System.Drawing.Point(106, 12);
            this.btnOpenPgk.Name = "btnOpenPgk";
            this.btnOpenPgk.Size = new System.Drawing.Size(105, 30);
            this.btnOpenPgk.TabIndex = 1;
            this.btnOpenPgk.Text = "从 PGK 提取";
            this.btnOpenPgk.UseVisualStyleBackColor = true;
            this.btnOpenPgk.Click += new System.EventHandler(this.btnOpenPgk_Click);
            //
            this.btnExportPng.Location = new System.Drawing.Point(217, 12);
            this.btnExportPng.Name = "btnExportPng";
            this.btnExportPng.Size = new System.Drawing.Size(95, 30);
            this.btnExportPng.TabIndex = 2;
            this.btnExportPng.Text = "导出全部 PNG";
            this.btnExportPng.UseVisualStyleBackColor = true;
            this.btnExportPng.Click += new System.EventHandler(this.btnExportPng_Click);
            //
            this.btnExportCharacter.Location = new System.Drawing.Point(318, 12);
            this.btnExportCharacter.Name = "btnExportCharacter";
            this.btnExportCharacter.Size = new System.Drawing.Size(95, 30);
            this.btnExportCharacter.TabIndex = 3;
            this.btnExportCharacter.Text = "合成角色";
            this.btnExportCharacter.UseVisualStyleBackColor = true;
            this.btnExportCharacter.Click += new System.EventHandler(this.btnExportCharacter_Click);
            //
            // 第二行按钮（右侧工具按钮）
            //
            this.btnExportCurrentFrame.Location = new System.Drawing.Point(419, 12);
            this.btnExportCurrentFrame.Name = "btnExportCurrentFrame";
            this.btnExportCurrentFrame.Size = new System.Drawing.Size(100, 30);
            this.btnExportCurrentFrame.TabIndex = 4;
            this.btnExportCurrentFrame.Text = "导出当前帧";
            this.btnExportCurrentFrame.UseVisualStyleBackColor = true;
            this.btnExportCurrentFrame.Click += new System.EventHandler(this.btnExportCurrentFrame_Click);
            //
            this.btnBatchExport.Location = new System.Drawing.Point(525, 12);
            this.btnBatchExport.Name = "btnBatchExport";
            this.btnBatchExport.Size = new System.Drawing.Size(95, 30);
            this.btnBatchExport.TabIndex = 6;
            this.btnBatchExport.Text = "批量导出库";
            this.btnBatchExport.UseVisualStyleBackColor = true;
            this.btnBatchExport.Click += new System.EventHandler(this.btnBatchExport_Click);
            //
            // 动画工具组
            //
            this.grpTools.Controls.Add(this.chkLoop);
            this.grpTools.Controls.Add(this.numFps);
            this.grpTools.Controls.Add(this.lblFps);
            this.grpTools.Controls.Add(this.btnExportGif);
            this.grpTools.Controls.Add(this.btnStopAnimation);
            this.grpTools.Controls.Add(this.btnPlayAnimation);
            this.grpTools.Location = new System.Drawing.Point(727, 8);
            this.grpTools.Name = "grpTools";
            this.grpTools.Size = new System.Drawing.Size(165, 52);
            this.grpTools.TabIndex = 7;
            this.grpTools.TabStop = false;
            this.grpTools.Text = "动画工具";
            //
            this.btnPlayAnimation.BackColor = System.Drawing.Color.LightGreen;
            this.btnPlayAnimation.Location = new System.Drawing.Point(6, 20);
            this.btnPlayAnimation.Name = "btnPlayAnimation";
            this.btnPlayAnimation.Size = new System.Drawing.Size(50, 26);
            this.btnPlayAnimation.TabIndex = 0;
            this.btnPlayAnimation.Text = "▶播放";
            this.btnPlayAnimation.UseVisualStyleBackColor = false;
            this.btnPlayAnimation.Click += new System.EventHandler(this.btnPlayAnimation_Click);
            //
            this.btnStopAnimation.BackColor = System.Drawing.Color.LightCoral;
            this.btnStopAnimation.Enabled = false;
            this.btnStopAnimation.Location = new System.Drawing.Point(60, 20);
            this.btnStopAnimation.Name = "btnStopAnimation";
            this.btnStopAnimation.Size = new System.Drawing.Size(50, 26);
            this.btnStopAnimation.TabIndex = 1;
            this.btnStopAnimation.Text = "■停止";
            this.btnStopAnimation.UseVisualStyleBackColor = false;
            this.btnStopAnimation.Click += new System.EventHandler(this.btnStopAnimation_Click);
            //
            this.btnExportGif.Location = new System.Drawing.Point(114, 20);
            this.btnExportGif.Name = "btnExportGif";
            this.btnExportGif.Size = new System.Drawing.Size(45, 26);
            this.btnExportGif.TabIndex = 2;
            this.btnExportGif.Text = "GIF";
            this.btnExportGif.UseVisualStyleBackColor = true;
            this.btnExportGif.Click += new System.EventHandler(this.btnExportGif_Click);
            //
            this.lblFps.AutoSize = true;
            this.lblFps.Location = new System.Drawing.Point(6, 48);
            this.lblFps.Name = "lblFps";
            this.lblFps.Size = new System.Drawing.Size(28, 16);
            this.lblFps.TabIndex = 3;
            this.lblFps.Text = "FPS:";
            //
            this.numFps.Location = new System.Drawing.Point(34, 45);
            this.numFps.Maximum = new decimal(new int[] { 60, 0, 0, 0 });
            this.numFps.Minimum = new decimal(new int[] { 1, 0, 0, 0 });
            this.numFps.Name = "numFps";
            this.numFps.Size = new System.Drawing.Size(42, 22);
            this.numFps.TabIndex = 4;
            this.numFps.Value = new decimal(new int[] { 8, 0, 0, 0 });
            //
            this.chkLoop.AutoSize = true;
            this.chkLoop.Checked = true;
            this.chkLoop.CheckState = System.Windows.Forms.CheckState.Checked;
            this.chkLoop.Location = new System.Drawing.Point(82, 47);
            this.chkLoop.Name = "chkLoop";
            this.chkLoop.Size = new System.Drawing.Size(56, 18);
            this.chkLoop.TabIndex = 5;
            this.chkLoop.Text = "循环";
            this.chkLoop.UseVisualStyleBackColor = true;
            //
            // 信息标签
            //
            this.lblInfo.AutoSize = true;
            this.lblInfo.Location = new System.Drawing.Point(10, 50);
            this.lblInfo.Name = "lblInfo";
            this.lblInfo.Size = new System.Drawing.Size(200, 18);
            this.lblInfo.TabIndex = 8;
            this.lblInfo.Text = "请打开 .atz 文件或从 PGK 包提取";
            //
            // 预览图片
            //
            this.picPreview.BackColor = System.Drawing.Color.FromArgb(((int)(((byte)(64)))), ((int)(((byte)(64)))), ((int)(((byte)(64)))));
            this.picPreview.BorderStyle = System.Windows.Forms.BorderStyle.FixedSingle;
            this.picPreview.Location = new System.Drawing.Point(10, 78);
            this.picPreview.Name = "picPreview";
            this.picPreview.Size = new System.Drawing.Size(380, 380);
            this.picPreview.SizeMode = System.Windows.Forms.PictureBoxSizeMode.Zoom;
            this.picPreview.TabIndex = 9;
            this.picPreview.TabStop = false;
            //
            // 帧列表
            //
            this.lstFrames.FormattingEnabled = true;
            this.lstFrames.IntegralHeight = false;
            this.lstFrames.ItemHeight = 17;
            this.lstFrames.Location = new System.Drawing.Point(398, 78);
            this.lstFrames.Name = "lstFrames";
            this.lstFrames.Size = new System.Drawing.Size(180, 380);
            this.lstFrames.TabIndex = 10;
            this.lstFrames.SelectedIndexChanged += new System.EventHandler(this.lstFrames_SelectedIndexChanged);
            //
            // PGK文件列表
            //
            this.lstPgkFiles.FormattingEnabled = true;
            this.lstPgkFiles.IntegralHeight = false;
            this.lstPgkFiles.ItemHeight = 17;
            this.lstPgkFiles.Location = new System.Drawing.Point(586, 78);
            this.lstPgkFiles.Name = "lstPgkFiles";
            this.lstPgkFiles.Size = new System.Drawing.Size(306, 380);
            this.lstPgkFiles.TabIndex = 11;
            this.lstPgkFiles.SelectedIndexChanged += new System.EventHandler(this.lstPgkFiles_SelectedIndexChanged);
            //
            // 帧滑块
            //
            this.trkFrame.Enabled = false;
            this.trkFrame.Location = new System.Drawing.Point(10, 466);
            this.trkFrame.Maximum = 0;
            this.trkFrame.Name = "trkFrame";
            this.trkFrame.Size = new System.Drawing.Size(882, 40);
            this.trkFrame.TabIndex = 12;
            this.trkFrame.Scroll += new System.EventHandler(this.trkFrame_Scroll);
            //
            // FormAtzViewer
            //
            this.AutoScaleDimensions = new System.Drawing.SizeF(7F, 15F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(910, 520);
            this.Controls.Add(this.trkFrame);
            this.Controls.Add(this.lstPgkFiles);
            this.Controls.Add(this.lstFrames);
            this.Controls.Add(this.picPreview);
            this.Controls.Add(this.lblInfo);
            this.Controls.Add(this.grpTools);
            this.Controls.Add(this.btnBatchExport);
            this.Controls.Add(this.btnExportCurrentFrame);
            this.Controls.Add(this.btnExportCharacter);
            this.Controls.Add(this.btnExportPng);
            this.Controls.Add(this.btnOpenPgk);
            this.Controls.Add(this.btnOpenAtz);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.Sizable;
            this.MaximizeBox = true;
            this.MinimizeBox = true;
            this.MinimumSize = new System.Drawing.Size(800, 450);
            this.Name = "FormAtzViewer";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "ATZ 资源预览与导出";
            ((System.ComponentModel.ISupportInitialize)(this.picPreview)).EndInit();
            ((System.ComponentModel.ISupportInitialize)(this.trkFrame)).EndInit();
            ((System.ComponentModel.ISupportInitialize)(this.numFps)).EndInit();
            this.grpTools.ResumeLayout(false);
            this.grpTools.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();
        }

        #endregion

        private System.Windows.Forms.Button btnOpenAtz;
        private System.Windows.Forms.Button btnOpenPgk;
        private System.Windows.Forms.Button btnExportPng;
        private System.Windows.Forms.Button btnExportCharacter;
        private System.Windows.Forms.Label lblInfo;
        private System.Windows.Forms.PictureBox picPreview;
        private System.Windows.Forms.ListBox lstFrames;
        private System.Windows.Forms.ListBox lstPgkFiles;
        private System.Windows.Forms.TrackBar trkFrame;
        private System.Windows.Forms.GroupBox grpTools;
        private System.Windows.Forms.Button btnPlayAnimation;
        private System.Windows.Forms.Button btnStopAnimation;
        private System.Windows.Forms.Button btnExportGif;
        private System.Windows.Forms.Label lblFps;
        private System.Windows.Forms.NumericUpDown numFps;
        private System.Windows.Forms.CheckBox chkLoop;
        private System.Windows.Forms.Button btnExportCurrentFrame;
        private System.Windows.Forms.Button btnBatchExport;
    }
}
