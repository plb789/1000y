using System;
using System.Text;
using System.Windows.Forms;
using WsDebugTool.Network;
using WsDebugTool.Utils;

namespace WsDebugTool
{
    public partial class FormMain : Form
    {
        private readonly WsClient _wsClient;

        public FormMain()
        {
            InitializeComponent();
            _wsClient = new WsClient();
            BindEvent();

            // 默认地址（本地网关）
            txtUrl.Text = "ws://127.0.0.1:8080/ws";
            // 预设常用协议号
            cboCmd.Items.AddRange(new object[]
            {
                "0x0001 心跳",
                "0x0003 移动",
                "0x0004 聊天"
            });
            cboCmd.SelectedIndex = 0;
        }

        #region 绑定事件
        private void BindEvent()
        {
            _wsClient.OnConnected += () => lblStatus.Text = "状态：已连接";
            _wsClient.OnDisconnected += () => lblStatus.Text = "状态：未连接";
            _wsClient.OnLog += msg => LogHelper.Append(txtLog, msg);
            _wsClient.OnRecvPacket += (cmd, body) =>
            {
                // 可扩展协议解析展示
            };
        }
        #endregion

        #region 按钮事件
        // 连接
        private void btnConnect_Click(object sender, EventArgs e)
        {
            string url = txtUrl.Text.Trim();
            if (string.IsNullOrEmpty(url))
            {
                MessageBox.Show("请输入WebSocket地址");
                return;
            }
            _ = _wsClient.Connect(url);
        }

        // 断开
        private void btnDisConnect_Click(object sender, EventArgs e)
        {
            _ = _wsClient.DisConnect();
        }

        // 自动心跳
        private void chkAutoHeart_CheckedChanged(object sender, EventArgs e)
        {
            _wsClient.SetAutoHeart(chkAutoHeart.Checked);
        }

        // 发送封包
        private async void btnSend_Click(object sender, EventArgs e)
        {
            if (!_wsClient.IsConnected)
            {
                MessageBox.Show("请先连接服务端");
                return;
            }

            // 解析协议号
            string cmdStr = cboCmd.Text.Split(' ')[0].Trim();
            if (!ushort.TryParse(cmdStr.TrimStart('0', 'x'), System.Globalization.NumberStyles.HexNumber, null, out ushort cmd))
            {
                MessageBox.Show("协议号格式错误");
                return;
            }

            // 解析消息体（支持纯文本/十六进制）
            string bodyStr = txtBody.Text.Trim();
            byte[] body = Array.Empty<byte>();
            if (!string.IsNullOrEmpty(bodyStr))
            {
                // 简易：文本直接转ASCII，也可扩展十六进制解析
                body = Encoding.UTF8.GetBytes(bodyStr);
            }

            await _wsClient.SendPacket(cmd, body);
        }

        // 清空日志
        private void btnClearLog_Click(object sender, EventArgs e)
        {
            LogHelper.ClearLog(txtLog);
        }
        #endregion
    }
}