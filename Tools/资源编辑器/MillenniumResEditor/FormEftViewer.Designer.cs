namespace MillenniumResEditor
{
    partial class FormEftViewer
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
            this.btnOpenEft = new System.Windows.Forms.Button();
            this.btnOpenPgk = new System.Windows.Forms.Button();
            this.btnExportPng = new System.Windows.Forms.Button();
            this.btnExportGif = new System.Windows.Forms.Button();
            this.lblInfo = new System.Windows.Forms.Label();
            this.picPreview = new System.Windows.Forms.PictureBox();
            this.lstFrames = new System.Windows.Forms.ListBox();
            this.lstPgkFiles = new System.Windows.Forms.ListBox();
            this.trkFrame = new System.Windows.Forms.TrackBar();
            this.grpTools = new System.Windows.Forms.GroupBox();
            this.btnPlayAnimation = new System.Windows.Forms.Button();
            this.btnStopAnimation = new System.Windows.Forms.Button();
            this.numFps = new System.Windows.Forms.NumericUpDown();
            this.lblFps = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.picPreview)).BeginInit();
            ((System.ComponentModel.ISupportInitialize)(this.trkFrame)).BeginInit();
            ((System.ComponentModel.ISupportInitialize)(this.numFps)).BeginInit();
            this.grpTools.SuspendLayout();
            this.SuspendLayout();
            //
            // btnOpenEft
            //
            this.btnOpenEft.Location = new System.Drawing.Point(10, 12);
            this.btnOpenEft.Name = "btnOpenEft";
            this.btnOpenEft.Size = new System.Drawing.Size(90, 30);
            this.btnOpenEft.TabIndex = 0;
            this.btnOpenEft.Text = "打开 .eft";
            this.btnOpenEft.UseVisualStyleBackColor = true;
            this.btnOpenEft.Click += new System.EventHandler(this.btnOpenEft_Click);
            //
            // btnOpenPgk
            //
            this.btnOpenPgk.Location = new System.Drawing.Point(106, 12);
            this.btnOpenPgk.Name = "btnOpenPgk";
            this.btnOpenPgk.Size = new System.Drawing.Size(110, 30);
            this.btnOpenPgk.TabIndex = 1;
            this.btnOpenPgk.Text = "从 PGK 提取";
            this.btnOpenPgk.UseVisualStyleBackColor = true;
            this.btnOpenPgk.Click += new System.EventHandler(this.btnOpenPgk_Click);
            //
            // btnExportPng
            //
            this.btnExportPng.Location = new System.Drawing.Point(222, 12);
            this.btnExportPng.Name = "btnExportPng";
            this.btnExportPng.Size = new System.Drawing.Size(100, 30);
            this.btnExportPng.TabIndex = 2;
            this.btnExportPng.Text = "导出全部 PNG";
            this.btnExportPng.UseVisualStyleBackColor = true;
            this.btnExportPng.Click += new System.EventHandler(this.btnExportPng_Click);
            //
            // btnExportGif
            //
            this.btnExportGif.Location = new System.Drawing.Point(328, 12);
            this.btnExportGif.Name = "btnExportGif";
            this.btnExportGif.Size = new System.Drawing.Size(90, 30);
            this.btnExportGif.TabIndex = 3;
            this.btnExportGif.Text = "导出 GIF";
            this.btnExportGif.UseVisualStyleBackColor = true;
            this.btnExportGif.Click += new System.EventHandler(this.btnExportGif_Click);
            //
            // grpTools
            //
            this.grpTools.Controls.Add(this.numFps);
            this.grpTools.Controls.Add(this.lblFps);
            this.grpTools.Controls.Add(this.btnStopAnimation);
            this.grpTools.Controls.Add(this.btnPlayAnimation);
            this.grpTools.Location = new System.Drawing.Point(424, 8);
            this.grpTools.Name = "grpTools";
            this.grpTools.Size = new System.Drawing.Size(220, 52);
            this.grpTools.TabIndex = 4;
            this.grpTools.TabStop = false;
            this.grpTools.Text = "动画工具";
            //
            // btnPlayAnimation
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
            // btnStopAnimation
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
            // lblFps
            //
            this.lblFps.AutoSize = true;
            this.lblFps.Location = new System.Drawing.Point(116, 25);
            this.lblFps.Name = "lblFps";
            this.lblFps.Size = new System.Drawing.Size(28, 16);
            this.lblFps.TabIndex = 2;
            this.lblFps.Text = "FPS:";
            //
            // numFps
            //
            this.numFps.Location = new System.Drawing.Point(144, 22);
            this.numFps.Maximum = new decimal(new int[] { 60, 0, 0, 0 });
            this.numFps.Minimum = new decimal(new int[] { 1, 0, 0, 0 });
            this.numFps.Name = "numFps";
            this.numFps.Size = new System.Drawing.Size(42, 22);
            this.numFps.TabIndex = 3;
            this.numFps.Value = new decimal(new int[] { 10, 0, 0, 0 });
            //
            // lblInfo
            //
            this.lblInfo.AutoSize = true;
            this.lblInfo.Location = new System.Drawing.Point(10, 48);
            this.lblInfo.Name = "lblInfo";
            this.lblInfo.Size = new System.Drawing.Size(200, 18);
            this.lblInfo.TabIndex = 5;
            this.lblInfo.Text = "请打开 .eft 文件或从 eft.pgk 包提取";
            //
            // picPreview
            //
            this.picPreview.BackColor = System.Drawing.Color.FromArgb(((int)(((byte)(64)))), ((int)(((byte)(64)))), ((int)(((byte)(64)))));
            this.picPreview.BorderStyle = System.Windows.Forms.BorderStyle.FixedSingle;
            this.picPreview.Location = new System.Drawing.Point(10, 75);
            this.picPreview.Name = "picPreview";
            this.picPreview.Size = new System.Drawing.Size(380, 380);
            this.picPreview.SizeMode = System.Windows.Forms.PictureBoxSizeMode.Zoom;
            this.picPreview.TabIndex = 6;
            this.picPreview.TabStop = false;
            //
            // lstFrames
            //
            this.lstFrames.FormattingEnabled = true;
            this.lstFrames.IntegralHeight = false;
            this.lstFrames.ItemHeight = 17;
            this.lstFrames.Location = new System.Drawing.Point(398, 75);
            this.lstFrames.Name = "lstFrames";
            this.lstFrames.Size = new System.Drawing.Size(180, 380);
            this.lstFrames.TabIndex = 7;
            this.lstFrames.SelectedIndexChanged += new System.EventHandler(this.lstFrames_SelectedIndexChanged);
            //
            // lstPgkFiles
            //
            this.lstPgkFiles.FormattingEnabled = true;
            this.lstPgkFiles.IntegralHeight = false;
            this.lstPgkFiles.ItemHeight = 17;
            this.lstPgkFiles.Location = new System.Drawing.Point(586, 75);
            this.lstPgkFiles.Name = "lstPgkFiles";
            this.lstPgkFiles.Size = new System.Drawing.Size(306, 380);
            this.lstPgkFiles.TabIndex = 8;
            this.lstPgkFiles.SelectedIndexChanged += new System.EventHandler(this.lstPgkFiles_SelectedIndexChanged);
            //
            // trkFrame
            //
            this.trkFrame.Enabled = false;
            this.trkFrame.Location = new System.Drawing.Point(10, 463);
            this.trkFrame.Maximum = 0;
            this.trkFrame.Name = "trkFrame";
            this.trkFrame.Size = new System.Drawing.Size(882, 40);
            this.trkFrame.TabIndex = 9;
            this.trkFrame.Scroll += new System.EventHandler(this.trkFrame_Scroll);
            //
            // FormEftViewer
            //
            this.AutoScaleDimensions = new System.Drawing.SizeF(7F, 15F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(910, 515);
            this.Controls.Add(this.trkFrame);
            this.Controls.Add(this.lstPgkFiles);
            this.Controls.Add(this.lstFrames);
            this.Controls.Add(this.picPreview);
            this.Controls.Add(this.lblInfo);
            this.Controls.Add(this.grpTools);
            this.Controls.Add(this.btnExportGif);
            this.Controls.Add(this.btnExportPng);
            this.Controls.Add(this.btnOpenPgk);
            this.Controls.Add(this.btnOpenEft);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.Sizable;
            this.MaximizeBox = true;
            this.MinimizeBox = true;
            this.MinimumSize = new System.Drawing.Size(800, 450);
            this.Name = "FormEftViewer";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "EFT 特效查看与导出";
            ((System.ComponentModel.ISupportInitialize)(this.picPreview)).EndInit();
            ((System.ComponentModel.ISupportInitialize)(this.trkFrame)).EndInit();
            ((System.ComponentModel.ISupportInitialize)(this.numFps)).EndInit();
            this.grpTools.ResumeLayout(false);
            this.grpTools.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();
        }

        #endregion

        private System.Windows.Forms.Button btnOpenEft;
        private System.Windows.Forms.Button btnOpenPgk;
        private System.Windows.Forms.Button btnExportPng;
        private System.Windows.Forms.Button btnExportGif;
        private System.Windows.Forms.Label lblInfo;
        private System.Windows.Forms.PictureBox picPreview;
        private System.Windows.Forms.ListBox lstFrames;
        private System.Windows.Forms.ListBox lstPgkFiles;
        private System.Windows.Forms.TrackBar trkFrame;
        private System.Windows.Forms.GroupBox grpTools;
        private System.Windows.Forms.Button btnPlayAnimation;
        private System.Windows.Forms.Button btnStopAnimation;
        private System.Windows.Forms.NumericUpDown numFps;
        private System.Windows.Forms.Label lblFps;
    }
}
