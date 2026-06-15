namespace MillenniumResEditor
{
    partial class FormMain
    {
        private System.ComponentModel.IContainer components = null;
        private System.Windows.Forms.Button btnOpenMap;
        private System.Windows.Forms.Button btnOpenSpr;
        private System.Windows.Forms.Button btnOpenDds;

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
            this.SuspendLayout();

            // btnOpenMap
            this.btnOpenMap.Location = new System.Drawing.Point(40, 40);
            this.btnOpenMap.Name = "btnOpenMap";
            this.btnOpenMap.Size = new System.Drawing.Size(140, 40);
            this.btnOpenMap.Text = "打开 .map 地图";
            this.btnOpenMap.UseVisualStyleBackColor = true;
            this.btnOpenMap.Click += new System.EventHandler(this.btnOpenMap_Click);

            // btnOpenSpr
            this.btnOpenSpr.Location = new System.Drawing.Point(220, 40);
            this.btnOpenSpr.Name = "btnOpenSpr";
            this.btnOpenSpr.Size = new System.Drawing.Size(140, 40);
            this.btnOpenSpr.Text = "打开 .spr 动画";
            this.btnOpenSpr.UseVisualStyleBackColor = true;
            this.btnOpenSpr.Click += new System.EventHandler(this.btnOpenSpr_Click);

            // btnOpenDds
            this.btnOpenDds.Location = new System.Drawing.Point(400, 40);
            this.btnOpenDds.Name = "btnOpenDds";
            this.btnOpenDds.Size = new System.Drawing.Size(140, 40);
            this.btnOpenDds.Text = "打开 .dds 贴图";
            this.btnOpenDds.UseVisualStyleBackColor = true;
            this.btnOpenDds.Click += new System.EventHandler(this.btnOpenDds_Click);

            // FormMain
            this.ClientSize = new System.Drawing.Size(620, 180);
            this.Controls.Add(this.btnOpenMap);
            this.Controls.Add(this.btnOpenSpr);
            this.Controls.Add(this.btnOpenDds);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.MaximizeBox = false;
            this.Name = "FormMain";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.ResumeLayout(false);
        }
        #endregion
    }
}