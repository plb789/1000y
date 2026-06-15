namespace MillenniumResEditor
{
    partial class FormSprViewer
    {
        private System.ComponentModel.IContainer components = null;
        private System.Windows.Forms.PictureBox picSpr;
        private System.Windows.Forms.Label lblFrameCount;
        private System.Windows.Forms.Label lblCurFrame;

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
            this.picSpr = new System.Windows.Forms.PictureBox();
            this.lblFrameCount = new System.Windows.Forms.Label();
            this.lblCurFrame = new System.Windows.Forms.Label();
            ((System.ComponentModel.ISupportInitialize)(this.picSpr)).BeginInit();
            this.SuspendLayout();

            // picSpr
            this.picSpr.Location = new System.Drawing.Point(20, 50);
            this.picSpr.Name = "picSpr";
            this.picSpr.Size = new System.Drawing.Size(400, 400);
            this.picSpr.SizeMode = System.Windows.Forms.PictureBoxSizeMode.CenterImage;

            // lblFrameCount
            this.lblFrameCount.AutoSize = true;
            this.lblFrameCount.Location = new System.Drawing.Point(20, 15);
            this.lblFrameCount.Name = "lblFrameCount";
            this.lblFrameCount.Text = "总帧数";

            // lblCurFrame
            this.lblCurFrame.AutoSize = true;
            this.lblCurFrame.Location = new System.Drawing.Point(200, 15);
            this.lblCurFrame.Name = "lblCurFrame";
            this.lblCurFrame.Text = "当前帧";

            // FormSprViewer
            this.ClientSize = new System.Drawing.Size(450, 480);
            this.Controls.Add(this.picSpr);
            this.Controls.Add(this.lblFrameCount);
            this.Controls.Add(this.lblCurFrame);
            this.Name = "FormSprViewer";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.FormClosed += new System.Windows.Forms.FormClosedEventHandler(this.FormSprViewer_FormClosed);
            ((System.ComponentModel.ISupportInitialize)(this.picSpr)).EndInit();
            this.ResumeLayout(false);
            this.PerformLayout();
        }
        #endregion
    }
}