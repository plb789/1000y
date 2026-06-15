namespace WsDebugTool
{
    partial class FormMain
    {
        private System.ComponentModel.IContainer components = null;
        private TextBox txtUrl;
        private Button btnConnect;
        private Button btnDisConnect;
        private CheckBox chkAutoHeart;
        private ComboBox cboCmd;
        private TextBox txtBody;
        private Button btnSend;
        private TextBox txtLog;
        private Label lblUrl;
        private Label lblCmd;
        private Label lblBody;
        private Label lblStatus;
        private Button btnClearLog;

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
            this.txtUrl = new System.Windows.Forms.TextBox();
            this.lblUrl = new System.Windows.Forms.Label();
            this.btnConnect = new System.Windows.Forms.Button();
            this.btnDisConnect = new System.Windows.Forms.Button();
            this.chkAutoHeart = new System.Windows.Forms.CheckBox();
            this.lblCmd = new System.Windows.Forms.Label();
            this.cboCmd = new System.Windows.Forms.ComboBox();
            this.lblBody = new System.Windows.Forms.Label();
            this.txtBody = new System.Windows.Forms.TextBox();
            this.btnSend = new System.Windows.Forms.Button();
            this.txtLog = new System.Windows.Forms.TextBox();
            this.lblStatus = new System.Windows.Forms.Label();
            this.btnClearLog = new System.Windows.Forms.Button();
            this.SuspendLayout();

            // lblUrl
            this.lblUrl.AutoSize = true;
            this.lblUrl.Location = new System.Drawing.Point(12, 15);
            this.lblUrl.Text = "WS地址：";

            // txtUrl
            this.txtUrl.Location = new System.Drawing.Point(70, 12);
            this.txtUrl.Size = new System.Drawing.Size(420, 21);

            // btnConnect
            this.btnConnect.Location = new System.Drawing.Point(500, 10);
            this.btnConnect.Size = new System.Drawing.Size(80, 25);
            this.btnConnect.Text = "连接";
            this.btnConnect.Click += new System.EventHandler(this.btnConnect_Click);

            // btnDisConnect
            this.btnDisConnect.Location = new System.Drawing.Point(590, 10);
            this.btnDisConnect.Size = new System.Drawing.Size(80, 25);
            this.btnDisConnect.Text = "断开";
            this.btnDisConnect.Click += new System.EventHandler(this.btnDisConnect_Click);

            // chkAutoHeart
            this.chkAutoHeart.Location = new System.Drawing.Point(12, 45);
            this.chkAutoHeart.Text = "自动心跳(10s)";
            this.chkAutoHeart.CheckedChanged += new System.EventHandler(this.chkAutoHeart_CheckedChanged);

            // lblCmd
            this.lblCmd.Location = new System.Drawing.Point(180, 45);
            this.lblCmd.Text = "协议号：";

            // cboCmd
            this.cboCmd.DropDownStyle = System.Windows.Forms.ComboBox.DropDownList;
            this.cboCmd.Location = new System.Drawing.Point(230, 42);
            this.cboCmd.Size = new System.Drawing.Size(140, 21);

            // lblBody
            this.lblBody.Location = new System.Drawing.Point(380, 45);
            this.lblBody.Text = "消息体：";

            // txtBody
            this.txtBody.Location = new System.Drawing.Point(430, 42);
            this.txtBody.Size = new System.Drawing.Size(140, 21);

            // btnSend
            this.btnSend.Location = new System.Drawing.Point(590, 40);
            this.btnSend.Size = new System.Drawing.Size(80, 25);
            this.btnSend.Text = "发送封包";
            this.btnSend.Click += new System.EventHandler(this.btnSend_Click);

            // btnClearLog
            this.btnClearLog.Location = new System.Drawing.Point(590, 70);
            this.btnClearLog.Size = new System.Drawing.Size(80, 25);
            this.btnClearLog.Text = "清空日志";
            this.btnClearLog.Click += new System.EventHandler(this.btnClearLog_Click);

            // lblStatus
            this.lblStatus.AutoSize = true;
            this.lblStatus.Location = new System.Drawing.Point(12, 75);
            this.lblStatus.Text = "状态：未连接";

            // txtLog
            this.txtLog.Location = new System.Drawing.Point(12, 100);
            this.txtLog.Size = new System.Drawing.Size(658, 320);
            this.txtLog.Multiline = true;
            this.txtLog.ScrollBars = System.Windows.Forms.ScrollBars.Vertical;
            this.txtLog.ReadOnly = true;

            // FormMain
            this.ClientSize = new System.Drawing.Size(684, 432);
            this.Text = "千年江湖 - WebSocket 调试工具 V1.0";
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.MaximizeBox = false;
            this.Controls.Add(this.lblUrl);
            this.Controls.Add(this.txtUrl);
            this.Controls.Add(this.btnConnect);
            this.Controls.Add(this.btnDisConnect);
            this.Controls.Add(this.chkAutoHeart);
            this.Controls.Add(this.lblCmd);
            this.Controls.Add(this.cboCmd);
            this.Controls.Add(this.lblBody);
            this.Controls.Add(this.txtBody);
            this.Controls.Add(this.btnSend);
            this.Controls.Add(this.btnClearLog);
            this.Controls.Add(this.lblStatus);
            this.Controls.Add(this.txtLog);
            this.ResumeLayout(false);
            this.PerformLayout();
        }
        #endregion
    }
}