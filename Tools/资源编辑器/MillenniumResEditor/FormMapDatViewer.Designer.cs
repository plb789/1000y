namespace MillenniumResEditor
{
    partial class FormMapDatViewer
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
            this.btnOpenMapDat = new System.Windows.Forms.Button();
            this.btnOpenMap = new System.Windows.Forms.Button();
            this.btnExtractAll = new System.Windows.Forms.Button();
            this.btnExportTmx = new System.Windows.Forms.Button();
            this.btnPreviewBlock = new System.Windows.Forms.Button();
            this.btnExportPng = new System.Windows.Forms.Button();
            this.btnLoadTileLib = new System.Windows.Forms.Button();
            this.lblInfo = new System.Windows.Forms.Label();
            this.lstBlocks = new System.Windows.Forms.ListBox();
            this.picPreview = new System.Windows.Forms.PictureBox();
            this.txtDetails = new System.Windows.Forms.RichTextBox();
            this.grpBlocks = new System.Windows.Forms.GroupBox();
            this.grpPreview = new System.Windows.Forms.GroupBox();
            this.grpDetails = new System.Windows.Forms.GroupBox();
            this.lblMapSize = new System.Windows.Forms.Label();
            this.lblBlockSize = new System.Windows.Forms.Label();
            this.lblCurrentBlock = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.picPreview)).BeginInit();
            this.grpBlocks.SuspendLayout();
            this.grpPreview.SuspendLayout();
            this.grpDetails.SuspendLayout();
            this.SuspendLayout();
            //
            // btnOpenMapDat
            //
            this.btnOpenMapDat.BackColor = System.Drawing.Color.LightGreen;
            this.btnOpenMapDat.Location = new System.Drawing.Point(10, 12);
            this.btnOpenMapDat.Name = "btnOpenMapDat";
            this.btnOpenMapDat.Size = new System.Drawing.Size(130, 35);
            this.btnOpenMapDat.TabIndex = 0;
            this.btnOpenMapDat.Text = "打开 MAP .dat";
            this.btnOpenMapDat.UseVisualStyleBackColor = false;
            this.btnOpenMapDat.Click += new System.EventHandler(this.btnOpenMapDat_Click);
            //
            // btnOpenMap
            //
            this.btnOpenMap.Location = new System.Drawing.Point(150, 12);
            this.btnOpenMap.Name = "btnOpenMap";
            this.btnOpenMap.Size = new System.Drawing.Size(120, 35);
            this.btnOpenMap.TabIndex = 1;
            this.btnOpenMap.Text = "打开 .map 文件";
            this.btnOpenMap.UseVisualStyleBackColor = true;
            this.btnOpenMap.Click += new System.EventHandler(this.btnOpenMap_Click);
            //
            // btnExtractAll
            //
            this.btnExtractAll.Location = new System.Drawing.Point(280, 12);
            this.btnExtractAll.Name = "btnExtractAll";
            this.btnExtractAll.Size = new System.Drawing.Size(100, 35);
            this.btnExtractAll.TabIndex = 2;
            this.btnExtractAll.Text = "全部解包";
            this.btnExtractAll.UseVisualStyleBackColor = true;
            this.btnExtractAll.Click += new System.EventHandler(this.btnExtractAll_Click);
            //
            // btnExportTmx
            //
            this.btnExportTmx.Location = new System.Drawing.Point(390, 12);
            this.btnExportTmx.Name = "btnExportTmx";
            this.btnExportTmx.Size = new System.Drawing.Size(100, 35);
            this.btnExportTmx.TabIndex = 3;
            this.btnExportTmx.Text = "导出 TMX";
            this.btnExportTmx.UseVisualStyleBackColor = true;
            this.btnExportTmx.Click += new System.EventHandler(this.btnExportTmx_Click);
            //
            // btnPreviewBlock
            //
            this.btnPreviewBlock.Location = new System.Drawing.Point(500, 12);
            this.btnPreviewBlock.Name = "btnPreviewBlock";
            this.btnPreviewBlock.Size = new System.Drawing.Size(100, 35);
            this.btnPreviewBlock.TabIndex = 4;
            this.btnPreviewBlock.Text = "预览选中块";
            this.btnPreviewBlock.UseVisualStyleBackColor = true;
            this.btnPreviewBlock.Click += new System.EventHandler(this.btnPreviewBlock_Click);
            //
            // btnExportPng
            //
            this.btnExportPng.BackColor = System.Drawing.Color.LightYellow;
            this.btnExportPng.Location = new System.Drawing.Point(610, 12);
            this.btnExportPng.Name = "btnExportPng";
            this.btnExportPng.Size = new System.Drawing.Size(110, 35);
            this.btnExportPng.TabIndex = 5;
            this.btnExportPng.Text = "导出地图PNG";
            this.btnExportPng.UseVisualStyleBackColor = false;
            this.btnExportPng.Click += new System.EventHandler(this.btnExportPng_Click);
            //
            // btnLoadTileLib
            //
            this.btnLoadTileLib.BackColor = System.Drawing.Color.LightCyan;
            this.btnLoadTileLib.Location = new System.Drawing.Point(730, 12);
            this.btnLoadTileLib.Name = "btnLoadTileLib";
            this.btnLoadTileLib.Size = new System.Drawing.Size(120, 35);
            this.btnLoadTileLib.TabIndex = 6;
            this.btnLoadTileLib.Text = "加载瓦片图集";
            this.btnLoadTileLib.UseVisualStyleBackColor = false;
            this.btnLoadTileLib.Click += new System.EventHandler(this.btnLoadTileLib_Click);
            //
            // lblInfo
            //
            this.lblInfo.AutoSize = true;
            this.lblInfo.Location = new System.Drawing.Point(10, 53);
            this.lblInfo.Name = "lblInfo";
            this.lblInfo.Size = new System.Drawing.Size(500, 18);
            this.lblInfo.TabIndex = 5;
            this.lblInfo.Text = "请打开 MAP .dat 资源包或 .map 地图文件";
            //
            // grpBlocks
            //
            this.grpBlocks.Controls.Add(this.lstBlocks);
            this.grpBlocks.Location = new System.Drawing.Point(10, 78);
            this.grpBlocks.Name = "grpBlocks";
            this.grpBlocks.Size = new System.Drawing.Size(280, 420);
            this.grpBlocks.TabIndex = 6;
            this.grpBlocks.TabStop = false;
            this.grpBlocks.Text = "地图块列表";
            //
            // lstBlocks
            //
            this.lstBlocks.FormattingEnabled = true;
            this.lstBlocks.IntegralHeight = false;
            this.lstBlocks.ItemHeight = 17;
            this.lstBlocks.Location = new System.Drawing.Point(10, 25);
            this.lstBlocks.Name = "lstBlocks";
            this.lstBlocks.Size = new System.Drawing.Size(260, 388);
            this.lstBlocks.TabIndex = 0;
            this.lstBlocks.SelectedIndexChanged += new System.EventHandler(this.lstBlocks_SelectedIndexChanged);
            //
            // grpPreview
            //
            this.grpPreview.Controls.Add(this.picPreview);
            this.grpPreview.Controls.Add(this.lblCurrentBlock);
            this.grpPreview.Location = new System.Drawing.Point(300, 78);
            this.grpPreview.Name = "grpPreview";
            this.grpPreview.Size = new System.Drawing.Size(320, 250);
            this.grpPreview.TabIndex = 7;
            this.grpPreview.TabStop = false;
            this.grpPreview.Text = "块预览";
            //
            // lblCurrentBlock
            //
            this.lblCurrentBlock.AutoSize = true;
            this.lblCurrentBlock.ForeColor = System.Drawing.Color.DarkBlue;
            this.lblCurrentBlock.Location = new System.Drawing.Point(10, 25);
            this.lblCurrentBlock.Name = "lblCurrentBlock";
            this.lblCurrentBlock.Size = new System.Drawing.Size(200, 20);
            this.lblCurrentBlock.TabIndex = 1;
            this.lblCurrentBlock.Text = "当前块: -";
            //
            // picPreview
            //
            this.picPreview.BackColor = System.Drawing.Color.FromArgb(((int)(((byte)(64)))), ((int)(((byte)(64)))), ((int)(((byte)(64)))));
            this.picPreview.BorderStyle = System.Windows.Forms.BorderStyle.FixedSingle;
            this.picPreview.Location = new System.Drawing.Point(10, 50);
            this.picPreview.Name = "picPreview";
            this.picPreview.Size = new System.Drawing.Size(300, 190);
            this.picPreview.SizeMode = System.Windows.Forms.PictureBoxSizeMode.Zoom;
            this.picPreview.TabIndex = 2;
            this.picPreview.TabStop = false;
            //
            // grpDetails
            //
            this.grpDetails.Controls.Add(this.lblMapSize);
            this.grpDetails.Controls.Add(this.lblBlockSize);
            this.grpDetails.Controls.Add(this.txtDetails);
            this.grpDetails.Location = new System.Drawing.Point(630, 78);
            this.grpDetails.Name = "grpDetails";
            this.grpDetails.Size = new System.Drawing.Size(270, 420);
            this.grpDetails.TabIndex = 8;
            this.grpDetails.TabStop = false;
            this.grpDetails.Text = "详细信息";
            //
            // lblMapSize
            //
            this.lblMapSize.AutoSize = true;
            this.lblMapSize.Font = new System.Drawing.Font("Microsoft YaHei UI", 9F, System.Drawing.FontStyle.Bold);
            this.lblMapSize.Location = new System.Drawing.Point(10, 25);
            this.lblMapSize.Name = "lblMapSize";
            this.lblMapSize.Size = new System.Drawing.Size(200, 20);
            this.lblMapSize.TabIndex = 1;
            this.lblMapSize.Text = "地图尺寸: - x - 格";
            //
            // lblBlockSize
            //
            this.lblBlockSize.AutoSize = true;
            this.lblBlockSize.Font = new System.Drawing.Font("Microsoft YaHei UI", 9F, System.Drawing.FontStyle.Bold);
            this.lblBlockSize.Location = new System.Drawing.Point(10, 48);
            this.lblBlockSize.Name = "lblBlockSize";
            this.lblBlockSize.Size = new System.Drawing.Size(200, 20);
            this.lblBlockSize.TabIndex = 2;
            this.lblBlockSize.Text = "块大小: - x -";
            //
            // txtDetails
            //
            this.txtDetails.Font = new System.Drawing.Font("Consolas", 9F);
            this.txtDetails.Location = new System.Drawing.Point(10, 75);
            this.txtDetails.Name = "txtDetails";
            this.txtDetails.ReadOnly = true;
            this.txtDetails.Size = new System.Drawing.Size(250, 335);
            this.txtDetails.TabIndex = 3;
            this.txtDetails.Text = "";
            //
            // FormMapDatViewer
            //
            this.AutoScaleDimensions = new System.Drawing.SizeF(7F, 15F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(920, 520);
            this.Controls.Add(this.grpDetails);
            this.Controls.Add(this.grpPreview);
            this.Controls.Add(this.grpBlocks);
            this.Controls.Add(this.lblInfo);
            this.Controls.Add(this.btnLoadTileLib);
            this.Controls.Add(this.btnExportPng);
            this.Controls.Add(this.btnPreviewBlock);
            this.Controls.Add(this.btnExportTmx);
            this.Controls.Add(this.btnExtractAll);
            this.Controls.Add(this.btnOpenMap);
            this.Controls.Add(this.btnOpenMapDat);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.Sizable;
            this.MaximizeBox = true;
            this.MinimizeBox = true;
            this.MinimumSize = new System.Drawing.Size(800, 450);
            this.Name = "FormMapDatViewer";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "MAP .DAT 地图包查看器";
            ((System.ComponentModel.ISupportInitialize)(this.picPreview)).EndInit();
            this.grpBlocks.ResumeLayout(false);
            this.grpPreview.ResumeLayout(false);
            this.grpPreview.PerformLayout();
            this.grpDetails.ResumeLayout(false);
            this.grpDetails.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();
        }

        #endregion

        private System.Windows.Forms.Button btnOpenMapDat;
        private System.Windows.Forms.Button btnOpenMap;
        private System.Windows.Forms.Button btnExtractAll;
        private System.Windows.Forms.Button btnExportTmx;
        private System.Windows.Forms.Button btnPreviewBlock;
        private System.Windows.Forms.Button btnExportPng;
        private System.Windows.Forms.Button btnLoadTileLib;
        private System.Windows.Forms.Label lblInfo;
        private System.Windows.Forms.ListBox lstBlocks;
        private System.Windows.Forms.PictureBox picPreview;
        private System.Windows.Forms.RichTextBox txtDetails;
        private System.Windows.Forms.GroupBox grpBlocks;
        private System.Windows.Forms.GroupBox grpPreview;
        private System.Windows.Forms.GroupBox grpDetails;
        private System.Windows.Forms.Label lblMapSize;
        private System.Windows.Forms.Label lblBlockSize;
        private System.Windows.Forms.Label lblCurrentBlock;
    }
}
