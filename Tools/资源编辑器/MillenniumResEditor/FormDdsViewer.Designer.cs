namespace MillenniumResEditor
{
    partial class FormDdsViewer
    {
        private System.ComponentModel.IContainer components = null;
        private System.Windows.Forms.PictureBox picDds;

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
            this.picDds = new System.Windows.Forms.PictureBox();
            ((System.ComponentModel.ISupportInitialize)(this.picDds)).BeginInit();
            this.SuspendLayout();

            // picDds
            this.picDds.Dock = System.Windows.Forms.DockStyle.Fill;
            this.picDds.Location = new System.Drawing.Point(0, 0);
            this.picDds.Name = "picDds";
            this.picDds.Size = new System.Drawing.Size(600, 500);
            this.picDds.SizeMode = System.Windows.Forms.PictureBoxSizeMode.Zoom;

            // FormDdsViewer
            this.ClientSize = new System.Drawing.Size(600, 500);
            this.Controls.Add(this.picDds);
            this.Name = "FormDdsViewer";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            ((System.ComponentModel.ISupportInitialize)(this.picDds)).EndInit();
            this.ResumeLayout(false);
        }
        #endregion
    }
}