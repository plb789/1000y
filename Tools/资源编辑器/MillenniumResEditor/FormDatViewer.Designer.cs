namespace MillenniumResEditor
{
    partial class FormDatViewer
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
            this.btnOpenDat = new System.Windows.Forms.Button();
            this.btnExtractAll = new System.Windows.Forms.Button();
            this.btnExtractSelected = new System.Windows.Forms.Button();
            this.lblInfo = new System.Windows.Forms.Label();
            this.lstFiles = new System.Windows.Forms.ListBox();
            this.txtPreview = new System.Windows.Forms.RichTextBox();
            this.grpFileList = new System.Windows.Forms.GroupBox();
            this.grpPreview = new System.Windows.Forms.GroupBox();
            this.lblFileName = new System.Windows.Forms.Label();
            this.lblFileSize = new System.Windows.Forms.Label();
            this.lblOffset = new System.Windows.Forms.Label();
            this.chkHexView = new System.Windows.Forms.CheckBox();
            this.btnSearch = new System.Windows.Forms.Button();
            this.txtSearch = new System.Windows.Forms.TextBox();
            this.picMapPreview = new System.Windows.Forms.PictureBox();
            this.btnPreviewMap = new System.Windows.Forms.Button();
            this.lblMapInfo = new System.Windows.Forms.Label();
            this.grpFileList.SuspendLayout();
            this.grpPreview.SuspendLayout();
            ((System.ComponentModel.ISupportInitialize)(this.picMapPreview)).BeginInit();
            this.SuspendLayout();
            //
            // btnOpenDat
            //
            this.btnOpenDat.BackColor = System.Drawing.Color.LightBlue;
            this.btnOpenDat.Location = new System.Drawing.Point(10, 12);
            this.btnOpenDat.Name = "btnOpenDat";
            this.btnOpenDat.Size = new System.Drawing.Size(120, 35);
            this.btnOpenDat.TabIndex = 0;
            this.btnOpenDat.Text = "打开 .dat 文件";
            this.btnOpenDat.UseVisualStyleBackColor = false;
            this.btnOpenDat.Click += new System.EventHandler(this.btnOpenDat_Click);
            //
            // btnExtractAll
            //
            this.btnExtractAll.Location = new System.Drawing.Point(140, 12);
            this.btnExtractAll.Name = "btnExtractAll";
            this.btnExtractAll.Size = new System.Drawing.Size(110, 35);
            this.btnExtractAll.TabIndex = 1;
            this.btnExtractAll.Text = "全部解包";
            this.btnExtractAll.UseVisualStyleBackColor = true;
            this.btnExtractAll.Click += new System.EventHandler(this.btnExtractAll_Click);
            //
            // btnExtractSelected
            //
            this.btnExtractSelected.Location = new System.Drawing.Point(260, 12);
            this.btnExtractSelected.Name = "btnExtractSelected";
            this.btnExtractSelected.Size = new System.Drawing.Size(120, 35);
            this.btnExtractSelected.TabIndex = 2;
            this.btnExtractSelected.Text = "解包选中项";
            this.btnExtractSelected.UseVisualStyleBackColor = true;
            this.btnExtractSelected.Click += new System.EventHandler(this.btnExtractSelected_Click);
            //
            // txtSearch
            //
            this.txtSearch.Location = new System.Drawing.Point(390, 17);
            this.txtSearch.Name = "txtSearch";
            this.txtSearch.Size = new System.Drawing.Size(200, 22);
            this.txtSearch.TabIndex = 3;
            this.txtSearch.PlaceholderText = "搜索文件名...";
            //
            // btnSearch
            //
            this.btnSearch.Location = new System.Drawing.Point(596, 15);
            this.btnSearch.Name = "btnSearch";
            this.btnSearch.Size = new System.Drawing.Size(60, 26);
            this.btnSearch.TabIndex = 4;
            this.btnSearch.Text = "搜索";
            this.btnSearch.UseVisualStyleBackColor = true;
            this.btnSearch.Click += new System.EventHandler(this.btnSearch_Click);
            //
            // btnPreviewMap
            //
            this.btnPreviewMap.BackColor = System.Drawing.Color.LightGreen;
            this.btnPreviewMap.Location = new System.Drawing.Point(664, 12);
            this.btnPreviewMap.Name = "btnPreviewMap";
            this.btnPreviewMap.Size = new System.Drawing.Size(100, 35);
            this.btnPreviewMap.TabIndex = 5;
            this.btnPreviewMap.Text = "地图预览";
            this.btnPreviewMap.UseVisualStyleBackColor = false;
            this.btnPreviewMap.Click += new System.EventHandler(this.btnPreviewMap_Click);
            //
            // lblMapInfo
            //
            this.lblMapInfo.AutoSize = true;
            this.lblMapInfo.ForeColor = System.Drawing.Color.DarkGreen;
            this.lblMapInfo.Location = new System.Drawing.Point(770, 22);
            this.lblMapInfo.Name = "lblMapInfo";
            this.lblMapInfo.Size = new System.Drawing.Size(120, 18);
            this.lblMapInfo.TabIndex = 6;
            this.lblMapInfo.Text = "";
            //
            // lblInfo
            //
            this.lblInfo.AutoSize = true;
            this.lblInfo.Location = new System.Drawing.Point(10, 53);
            this.lblInfo.Name = "lblInfo";
            this.lblInfo.Size = new System.Drawing.Size(400, 18);
            this.lblInfo.TabIndex = 5;
            this.lblInfo.Text = "请打开 .dat 资源包文件";
            //
            // grpFileList
            //
            this.grpFileList.Controls.Add(this.lstFiles);
            this.grpFileList.Location = new System.Drawing.Point(10, 78);
            this.grpFileList.Name = "grpFileList";
            this.grpFileList.Size = new System.Drawing.Size(300, 420);
            this.grpFileList.TabIndex = 6;
            this.grpFileList.TabStop = false;
            this.grpFileList.Text = "文件列表";
            //
            // lstFiles
            //
            this.lstFiles.FormattingEnabled = true;
            this.lstFiles.IntegralHeight = false;
            this.lstFiles.ItemHeight = 17;
            this.lstFiles.Location = new System.Drawing.Point(10, 25);
            this.lstFiles.Name = "lstFiles";
            this.lstFiles.Size = new System.Drawing.Size(280, 388);
            this.lstFiles.TabIndex = 0;
            this.lstFiles.SelectedIndexChanged += new System.EventHandler(this.lstFiles_SelectedIndexChanged);
            //
            // grpPreview
            //
            this.grpPreview.Controls.Add(this.picMapPreview);
            this.grpPreview.Controls.Add(this.chkHexView);
            this.grpPreview.Controls.Add(this.lblOffset);
            this.grpPreview.Controls.Add(this.lblFileSize);
            this.grpPreview.Controls.Add(this.lblFileName);
            this.grpPreview.Controls.Add(this.txtPreview);
            this.grpPreview.Location = new System.Drawing.Point(320, 78);
            this.grpPreview.Name = "grpPreview";
            this.grpPreview.Size = new System.Drawing.Size(570, 420);
            this.grpPreview.TabIndex = 7;
            this.grpPreview.TabStop = false;
            this.grpPreview.Text = "数据预览";
            //
            // lblFileName
            //
            this.lblFileName.AutoSize = true;
            this.lblFileName.Font = new System.Drawing.Font("Microsoft YaHei UI", 9F, System.Drawing.FontStyle.Bold);
            this.lblFileName.ForeColor = System.Drawing.Color.DarkBlue;
            this.lblFileName.Location = new System.Drawing.Point(10, 28);
            this.lblFileName.Name = "lblFileName";
            this.lblFileName.Size = new System.Drawing.Size(300, 20);
            this.lblFileName.TabIndex = 1;
            this.lblFileName.Text = "文件名: -";
            //
            // lblFileSize
            //
            this.lblFileSize.AutoSize = true;
            this.lblFileSize.Location = new System.Drawing.Point(10, 50);
            this.lblFileSize.Name = "lblFileSize";
            this.lblFileSize.Size = new System.Drawing.Size(200, 18);
            this.lblFileSize.TabIndex = 2;
            this.lblFileSize.Text = "大小: - 字节";
            //
            // lblOffset
            //
            this.lblOffset.AutoSize = true;
            this.lblOffset.Location = new System.Drawing.Point(10, 72);
            this.lblOffset.Name = "lblOffset";
            this.lblOffset.Size = new System.Drawing.Size(200, 18);
            this.lblOffset.TabIndex = 3;
            this.lblOffset.Text = "偏移: -";
            //
            // chkHexView
            //
            this.chkHexView.AutoSize = true;
            this.chkHexView.Checked = true;
            this.chkHexView.CheckState = System.Windows.Forms.CheckState.Checked;
            this.chkHexView.Location = new System.Drawing.Point(480, 50);
            this.chkHexView.Name = "chkHexView";
            this.chkHexView.Size = new System.Drawing.Size(80, 20);
            this.chkHexView.TabIndex = 4;
            this.chkHexView.Text = "十六进制";
            this.chkHexView.UseVisualStyleBackColor = true;
            this.chkHexView.CheckedChanged += new System.EventHandler(this.chkHexView_CheckedChanged);
            //
            // txtPreview
            //
            this.txtPreview.Font = new System.Drawing.Font("Consolas", 9F);
            this.txtPreview.Location = new System.Drawing.Point(10, 95);
            this.txtPreview.Name = "txtPreview";
            this.txtPreview.ReadOnly = true;
            this.txtPreview.Size = new System.Drawing.Size(280, 200);
            this.txtPreview.TabIndex = 5;
            this.txtPreview.Text = "";
            //
            // picMapPreview
            //
            this.picMapPreview.BackColor = System.Drawing.Color.FromArgb(((int)(((byte)(32)))), ((int)(((byte)(32)))), ((int)(((byte)(32)))));
            this.picMapPreview.BorderStyle = System.Windows.Forms.BorderStyle.FixedSingle;
            this.picMapPreview.Location = new System.Drawing.Point(300, 95);
            this.picMapPreview.Name = "picMapPreview";
            this.picMapPreview.Size = new System.Drawing.Size(260, 260);
            this.picMapPreview.SizeMode = System.Windows.Forms.PictureBoxSizeMode.Zoom;
            this.picMapPreview.TabIndex = 6;
            this.picMapPreview.TabStop = false;
            //
            // FormDatViewer
            //
            this.AutoScaleDimensions = new System.Drawing.SizeF(7F, 15F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(910, 520);
            this.Controls.Add(this.grpPreview);
            this.Controls.Add(this.grpFileList);
            this.Controls.Add(this.lblInfo);
            this.Controls.Add(this.lblMapInfo);
            this.Controls.Add(this.btnPreviewMap);
            this.Controls.Add(this.btnSearch);
            this.Controls.Add(this.txtSearch);
            this.Controls.Add(this.btnExtractSelected);
            this.Controls.Add(this.btnExtractAll);
            this.Controls.Add(this.btnOpenDat);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.Sizable;
            this.MaximizeBox = true;
            this.MinimizeBox = true;
            this.MinimumSize = new System.Drawing.Size(800, 450);
            this.Name = "FormDatViewer";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = ".DAT 资源包查看器";
            this.grpFileList.ResumeLayout(false);
            this.grpPreview.ResumeLayout(false);
            this.grpPreview.PerformLayout();
            this.ResumeLayout(false);
            this.PerformLayout();
        }

        #endregion

        private System.Windows.Forms.Button btnOpenDat;
        private System.Windows.Forms.Button btnExtractAll;
        private System.Windows.Forms.Button btnExtractSelected;
        private System.Windows.Forms.Label lblInfo;
        private System.Windows.Forms.ListBox lstFiles;
        private System.Windows.Forms.RichTextBox txtPreview;
        private System.Windows.Forms.GroupBox grpFileList;
        private System.Windows.Forms.GroupBox grpPreview;
        private System.Windows.Forms.Label lblFileName;
        private System.Windows.Forms.Label lblFileSize;
        private System.Windows.Forms.Label lblOffset;
        private System.Windows.Forms.CheckBox chkHexView;
        private System.Windows.Forms.Button btnSearch;
        private System.Windows.Forms.TextBox txtSearch;
        private System.Windows.Forms.PictureBox picMapPreview;
        private System.Windows.Forms.Button btnPreviewMap;
        private System.Windows.Forms.Label lblMapInfo;
    }
}
