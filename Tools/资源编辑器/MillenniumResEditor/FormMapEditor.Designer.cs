namespace MillenniumResEditor
{
    partial class FormMapEditor
    {
        private System.ComponentModel.IContainer components = null;
        private System.Windows.Forms.PictureBox picMap;
        private System.Windows.Forms.Label lblMapInfo;
        private System.Windows.Forms.Button btnSave;
        private System.Windows.Forms.Button btnSetBlock;
        private System.Windows.Forms.Button btnSetWalk;
        private System.Windows.Forms.Label lblCurAttr;

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
            this.picMap = new System.Windows.Forms.PictureBox();
            this.lblMapInfo = new System.Windows.Forms.Label();
            this.btnSave = new System.Windows.Forms.Button();
            this.btnSetBlock = new System.Windows.Forms.Button();
            this.btnSetWalk = new System.Windows.Forms.Button();
            this.lblCurAttr = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.picMap)).BeginInit();
            this.SuspendLayout();

            // picMap
            this.picMap.Location = new System.Drawing.Point(12, 60);
            this.picMap.Name = "picMap";
            this.picMap.Size = new System.Drawing.Size(800, 500);
            this.picMap.SizeMode = System.Windows.Forms.PictureBoxSizeMode.AutoSize;
            this.picMap.MouseDown += new System.Windows.Forms.MouseEventHandler(this.picMap_MouseDown);

            // lblMapInfo
            this.lblMapInfo.AutoSize = true;
            this.lblMapInfo.Location = new System.Drawing.Point(12, 15);
            this.lblMapInfo.Name = "lblMapInfo";
            this.lblMapInfo.Text = "地图信息";

            // btnSetWalk
            this.btnSetWalk.Location = new System.Drawing.Point(200, 12);
            this.btnSetWalk.Name = "btnSetWalk";
            this.btnSetWalk.Size = new System.Drawing.Size(90, 30);
            this.btnSetWalk.Text = "设为通行";
            this.btnSetWalk.Click += new System.EventHandler(this.btnSetWalk_Click);

            // btnSetBlock
            this.btnSetBlock.Location = new System.Drawing.Point(300, 12);
            this.btnSetBlock.Name = "btnSetBlock";
            this.btnSetBlock.Size = new System.Drawing.Size(90, 30);
            this.btnSetBlock.Text = "设为阻挡";
            this.btnSetBlock.Click += new System.EventHandler(this.btnSetBlock_Click);

            // btnSave
            this.btnSave.Location = new System.Drawing.Point(400, 12);
            this.btnSave.Name = "btnSave";
            this.btnSave.Size = new System.Drawing.Size(90, 30);
            this.btnSave.Text = "保存地图";
            this.btnSave.Click += new System.EventHandler(this.btnSave_Click);

            // lblCurAttr
            this.lblCurAttr.AutoSize = true;
            this.lblCurAttr.Location = new System.Drawing.Point(500, 20);
            this.lblCurAttr.Name = "lblCurAttr";
            this.lblCurAttr.Text = "当前属性：通行";

            // FormMapEditor
            this.ClientSize = new System.Drawing.Size(830, 580);
            this.Controls.Add(this.picMap);
            this.Controls.Add(this.lblMapInfo);
            this.Controls.Add(this.btnSetWalk);
            this.Controls.Add(this.btnSetBlock);
            this.Controls.Add(this.btnSave);
            this.Controls.Add(this.lblCurAttr);
            this.Name = "FormMapEditor";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            ((System.ComponentModel.ISupportInitialize)(this.picMap)).EndInit();
            this.ResumeLayout(false);
            this.PerformLayout();
        }
        #endregion
    }
}