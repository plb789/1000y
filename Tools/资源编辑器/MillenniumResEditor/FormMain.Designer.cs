namespace MillenniumResEditor
{
    partial class FormMain
    {
        private System.ComponentModel.IContainer components = null;
        private System.Windows.Forms.Button btnOpenMap;
        private System.Windows.Forms.Button btnOpenSpr;
        private System.Windows.Forms.Button btnOpenDds;
        private System.Windows.Forms.Button btnOpenAtz;
        private System.Windows.Forms.Button btnOpenEft;
        private System.Windows.Forms.Button btnOpenDat;
        private System.Windows.Forms.Button btnOpenMapDat;

        protected override void Dispose(bool disposing)
        {
            if (disposing && (components != null))
            {
                components.Dispose();
            }
            base.Dispose(disposing);
        }

        #region Windows 窗体设计器生成的代码
        private void InitializeComponent()
        {
            this.btnOpenMap = new System.Windows.Forms.Button();
            this.btnOpenSpr = new System.Windows.Forms.Button();
            this.btnOpenDds = new System.Windows.Forms.Button();
            this.btnOpenAtz = new System.Windows.Forms.Button();
            this.btnOpenEft = new System.Windows.Forms.Button();
            this.btnOpenDat = new System.Windows.Forms.Button();
            this.btnOpenMapDat = new System.Windows.Forms.Button();
            this.SuspendLayout();

            // btnOpenMap
            this.btnOpenMap.Location = new System.Drawing.Point(20, 15);
            this.btnOpenMap.Name = "btnOpenMap";
            this.btnOpenMap.Size = new System.Drawing.Size(120, 38);
            this.btnOpenMap.Text = "打开 .map 地图";
            this.btnOpenMap.UseVisualStyleBackColor = true;
            this.btnOpenMap.Click += new System.EventHandler(this.btnOpenMap_Click);

            // btnOpenSpr
            this.btnOpenSpr.Location = new System.Drawing.Point(155, 15);
            this.btnOpenSpr.Name = "btnOpenSpr";
            this.btnOpenSpr.Size = new System.Drawing.Size(120, 38);
            this.btnOpenSpr.Text = "打开 .spr 动画";
            this.btnOpenSpr.UseVisualStyleBackColor = true;
            this.btnOpenSpr.Click += new System.EventHandler(this.btnOpenSpr_Click);

            // btnOpenDds
            this.btnOpenDds.Location = new System.Drawing.Point(290, 15);
            this.btnOpenDds.Name = "btnOpenDds";
            this.btnOpenDds.Size = new System.Drawing.Size(120, 38);
            this.btnOpenDds.Text = "打开 .dds 贴图";
            this.btnOpenDds.UseVisualStyleBackColor = true;
            this.btnOpenDds.Click += new System.EventHandler(this.btnOpenDds_Click);

            // btnOpenAtz
            this.btnOpenAtz.Location = new System.Drawing.Point(425, 15);
            this.btnOpenAtz.Name = "btnOpenAtz";
            this.btnOpenAtz.Size = new System.Drawing.Size(120, 38);
            this.btnOpenAtz.Text = "打开 .atz 动画";
            this.btnOpenAtz.UseVisualStyleBackColor = true;
            this.btnOpenAtz.Click += new System.EventHandler(this.btnOpenAtz_Click);

            // btnOpenEft (新增)
            this.btnOpenEft.BackColor = System.Drawing.Color.LightGreen;
            this.btnOpenEft.Font = new System.Drawing.Font("Microsoft YaHei UI", 9F, System.Drawing.FontStyle.Bold);
            this.btnOpenEft.Location = new System.Drawing.Point(20, 65);
            this.btnOpenEft.Name = "btnOpenEft";
            this.btnOpenEft.Size = new System.Drawing.Size(120, 38);
            this.btnOpenEft.Text = "EFT 特效";
            this.btnOpenEft.UseVisualStyleBackColor = false;
            this.btnOpenEft.Click += new System.EventHandler(this.btnOpenEft_Click);

            // btnOpenDat (新增)
            this.btnOpenDat.BackColor = System.Drawing.Color.LightBlue;
            this.btnOpenDat.Font = new System.Drawing.Font("Microsoft YaHei UI", 9F, System.Drawing.FontStyle.Bold);
            this.btnOpenDat.Location = new System.Drawing.Point(155, 65);
            this.btnOpenDat.Name = "btnOpenDat";
            this.btnOpenDat.Size = new System.Drawing.Size(120, 38);
            this.btnOpenDat.Text = "DAT 解包";
            this.btnOpenDat.UseVisualStyleBackColor = false;
            this.btnOpenDat.Click += new System.EventHandler(this.btnOpenDat_Click);

            // btnOpenMapDat (新增)
            this.btnOpenMapDat.BackColor = System.Drawing.Color.LightYellow;
            this.btnOpenMapDat.Font = new System.Drawing.Font("Microsoft YaHei UI", 9F, System.Drawing.FontStyle.Bold);
            this.btnOpenMapDat.Location = new System.Drawing.Point(290, 65);
            this.btnOpenMapDat.Name = "btnOpenMapDat";
            this.btnOpenMapDat.Size = new System.Drawing.Size(130, 38);
            this.btnOpenMapDat.Text = "MAP DAT 包";
            this.btnOpenMapDat.UseVisualStyleBackColor = false;
            this.btnOpenMapDat.Click += new System.EventHandler(this.btnOpenMapDat_Click);

            // FormMain
            this.ClientSize = new System.Drawing.Size(580, 125);
            this.Controls.Add(this.btnOpenMapDat);
            this.Controls.Add(this.btnOpenDat);
            this.Controls.Add(this.btnOpenEft);
            this.Controls.Add(this.btnOpenAtz);
            this.Controls.Add(this.btnOpenMap);
            this.Controls.Add(this.btnOpenSpr);
            this.Controls.Add(this.btnOpenDds);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.MaximizeBox = false;
            this.MinimizeBox = true;
            this.Name = "FormMain";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.ResumeLayout(false);
        }
        #endregion
    }
}