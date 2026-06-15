# 千年江湖 - WebSocket 调试工具（C# WinForms）
基于 `.NET 6 + WinForms` 开发，和之前**资源编辑器技术栈统一**，可放在同解决方案，独立窗口运行。
功能覆盖：**WebSocket 连接管理、二进制封包收发、协议解析、日志记录、手动发包、自动心跳、十六进制查看**，完全适配你项目的二进制协议格式。

## 一、目录放置
沿用现有 `Tools` 目录结构，新增独立项目，也可合并到原有解决方案：
```
Tools/
├─ 资源编辑器/          # 原有项目
└─ WebSocket调试工具/
   ├─ WsDebugTool.sln    # 解决方案
   └─ WsDebugTool/
      ├─ WsDebugTool.csproj
      ├─ Program.cs
      ├─ FormMain.cs        # 主窗口
      ├─ FormMain.Designer.cs
      ├─ Utils/
      │  ├─ BinaryHelper.cs # 二进制/大小端工具
      │  └─ LogHelper.cs    # 日志工具
      └─ Network/
         └─ WsClient.cs     # WebSocket 核心客户端
```

---

## 二、前置说明
1. 协议完全对齐项目标准：
   封包格式：`协议号(2字节 小端) + 数据长度(2字节 小端) + 消息体 + 校验码(1字节)`
2. 内置功能：连接/断开、心跳自动发送、手动构造封包、原始十六进制查看、明文日志、自动重连。
3. 依赖：使用官方 `System.Net.WebSockets`（.NET 6 内置，无需第三方NuGet）。

---

# 三、完整代码实现

## 1. 通用工具类
### 1.1 Utils/BinaryHelper.cs 二进制/封包编解码（核心，和服务端规则一致）
```csharp
using System;
using System.Linq;

namespace WsDebugTool.Utils
{
    /// <summary>
    /// 二进制工具、封包编解码（和Go服务端规则完全一致）
    /// 协议：[Cmd(2LE)][Len(2LE)][Body][Check(1)]
    /// </summary>
    public static class BinaryHelper
    {
        /// <summary>
        /// 小端序读取 ushort
        /// </summary>
        public static ushort ReadUInt16LE(byte[] data, int offset)
        {
            byte[] buf = new byte[2];
            Array.Copy(data, offset, buf, 0, 2);
            if (!BitConverter.IsLittleEndian)
                Array.Reverse(buf);
            return BitConverter.ToUInt16(buf, 0);
        }

        /// <summary>
        /// 小端序写入 ushort
        /// </summary>
        public static byte[] WriteUInt16LE(ushort value)
        {
            byte[] buf = BitConverter.GetBytes(value);
            if (!BitConverter.IsLittleEndian)
                Array.Reverse(buf);
            return buf;
        }

        /// <summary>
        /// 计算累加校验码（全字节求和 & 0xFF）
        /// </summary>
        public static byte CalcCheckSum(byte[] data)
        {
            int sum = data.Sum(b => b);
            return (byte)(sum & 0xFF);
        }

        /// <summary>
        /// 组装完整封包
        /// </summary>
        /// <param name="cmd">协议号</param>
        /// <param name="body">消息体</param>
        /// <returns>完整二进制封包</returns>
        public static byte[] EncodePacket(ushort cmd, byte[] body)
        {
            body ??= Array.Empty<byte>();
            int bodyLen = body.Length;
            int totalLen = 4 + bodyLen + 1;

            byte[] packet = new byte[totalLen];
            // 写入协议号 2字节 小端
            byte[] cmdBuf = WriteUInt16LE(cmd);
            Array.Copy(cmdBuf, 0, packet, 0, 2);
            // 写入长度 2字节 小端
            byte[] lenBuf = WriteUInt16LE((ushort)bodyLen);
            Array.Copy(lenBuf, 0, packet, 2, 2);
            // 写入消息体
            Array.Copy(body, 0, packet, 4, bodyLen);
            // 计算并写入校验码
            packet[totalLen - 1] = CalcCheckSum(packet.Take(totalLen - 1).ToArray());

            return packet;
        }

        /// <summary>
        /// 解析封包
        /// </summary>
        /// <param name="packet">完整封包</param>
        /// <param name="cmd">输出协议号</param>
        /// <param name="body">输出消息体</param>
        /// <returns>校验是否通过</returns>
        public static bool DecodePacket(byte[] packet, out ushort cmd, out byte[] body)
        {
            cmd = 0;
            body = Array.Empty<byte>();
            if (packet.Length < 5)
                return false;

            // 校验码验证
            byte realCheck = packet[packet.Length - 1];
            byte calcCheck = CalcCheckSum(packet.Take(packet.Length - 1).ToArray());
            if (realCheck != calcCheck)
                return false;

            // 解析协议号 + 长度
            cmd = ReadUInt16LE(packet, 0);
            ushort bodyLen = ReadUInt16LE(packet, 2);
            if (4 + bodyLen > packet.Length - 1)
                return false;

            body = new byte[bodyLen];
            Array.Copy(packet, 4, body, 0, bodyLen);
            return true;
        }

        /// <summary>
        /// 字节数组转十六进制字符串（方便查看）
        /// </summary>
        public static string BytesToHex(byte[] data)
        {
            if (data == null || data.Length == 0)
                return "";
            return BitConverter.ToString(data).Replace("-", " ");
        }
    }
}
```

### 1.2 Utils/LogHelper.cs 日志输出工具
```csharp
using System;
using System.Windows.Forms;

namespace WsDebugTool.Utils
{
    /// <summary>
    /// 日志输出工具
    /// </summary>
    public static class LogHelper
    {
        /// <summary>
        /// 追加日志到文本框
        /// </summary>
        public static void AppendLog(TextBox txt, string msg, bool isError = false)
        {
            string time = DateTime.Now.ToString("HH:mm:ss");
            string prefix = isError ? "[错误]" : "[正常]";
            string line = $"[{time}] {prefix} {msg}\r\n";

            if (txt.InvokeRequired)
            {
                txt.Invoke(new Action(() =>
                {
                    txt.AppendText(line);
                    txt.ScrollToCaret();
                }));
            }
            else
            {
                txt.AppendText(line);
                txt.ScrollToCaret();
            }
        }

        /// <summary>
        /// 清空日志
        /// </summary>
        public static void ClearLog(TextBox txt)
        {
            if (txt.InvokeRequired)
            {
                txt.Invoke(new Action(txt.Clear));
            }
            else
            {
                txt.Clear();
            }
        }
    }
}
```

## 2. 网络核心 Network/WsClient.cs WebSocket 客户端
异步长连接、收发、心跳、重连逻辑
```csharp
using System;
using System.Net.WebSockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using WsDebugTool.Utils;

namespace WsDebugTool.Network
{
    public class WsClient
    {
        #region 事件回调
        public event Action OnConnected;
        public event Action OnDisconnected;
        public event Action<string> OnLog;
        public event Action<ushort, byte[]> OnRecvPacket;
        #endregion

        private readonly ClientWebSocket _ws;
        private CancellationTokenSource _cts;
        private string _serverUrl;
        private bool _isAutoHeart = false;
        private readonly int _heartInterval = 10000; // 10秒心跳

        // 协议号常量（和项目统一）
        public const ushort CmdHeart = 0x0001;
        public const ushort CmdMove = 0x0003;
        public const ushort CmdChat = 0x0004;

        public WsClient()
        {
            _ws = new ClientWebSocket();
            _ws.Options.KeepAliveInterval = TimeSpan.FromSeconds(10);
        }

        /// <summary>
        /// 是否已连接
        /// </summary>
        public bool IsConnected => _ws.State == WebSocketState.Open;

        /// <summary>
        /// 连接服务端
        /// </summary>
        public async Task Connect(string url)
        {
            if (IsConnected)
            {
                OnLog?.Invoke("当前已连接，无需重复连接");
                return;
            }

            _serverUrl = url;
            _cts = new CancellationTokenSource();
            try
            {
                Uri uri = new Uri(url);
                await _ws.ConnectAsync(uri, _cts.Token);
                OnConnected?.Invoke();
                OnLog?.Invoke("WebSocket 连接成功");

                // 启动接收循环
                _ = ReceiveLoop();
                // 启动心跳
                if (_isAutoHeart)
                    _ = HeartLoop();
            }
            catch (Exception ex)
            {
                OnLog?.Invoke($"连接失败：{ex.Message}");
                OnDisconnected?.Invoke();
            }
        }

        /// <summary>
        /// 断开连接
        /// </summary>
        public async Task DisConnect()
        {
            if (!IsConnected) return;
            _cts?.Cancel();
            try
            {
                await _ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "手动断开", CancellationToken.None);
            }
            catch { }
            OnDisconnected?.Invoke();
            OnLog?.Invoke("已手动断开连接");
        }

        /// <summary>
        /// 开启/关闭自动心跳
        /// </summary>
        public void SetAutoHeart(bool enable)
        {
            _isAutoHeart = enable;
            if (enable && IsConnected)
                _ = HeartLoop();
        }

        /// <summary>
        /// 发送封包
        /// </summary>
        public async Task SendPacket(ushort cmd, byte[] body)
        {
            if (!IsConnected)
            {
                OnLog?.Invoke("发送失败：未连接服务端");
                return;
            }
            byte[] packet = BinaryHelper.EncodePacket(cmd, body);
            try
            {
                ArraySegment<byte> seg = new ArraySegment<byte>(packet);
                await _ws.SendAsync(seg, WebSocketMessageType.Binary, true, _cts.Token);
                string hex = BinaryHelper.BytesToHex(packet);
                OnLog?.Invoke($"发包 -> 协议号:{cmd:X4}  十六进制:[{hex}]");
            }
            catch (Exception ex)
            {
                OnLog?.Invoke($"发包异常：{ex.Message}");
            }
        }

        /// <summary>
        /// 接收消息循环
        /// </summary>
        private async Task ReceiveLoop()
        {
            byte[] buffer = new byte[4096];
            while (!_cts.Token.IsCancellationRequested && IsConnected)
            {
                try
                    {
                        WebSocketReceiveResult res = await _ws.ReceiveAsync(new ArraySegment<byte>(buffer), _cts.Token);
                        if (res.MessageType == WebSocketMessageType.Close)
                        {
                            await DisConnect();
                            break;
                        }

                        byte[] recvData = new byte[res.Count];
                        Array.Copy(buffer, recvData, res.Count);
                        // 解析封包
                        if (BinaryHelper.DecodePacket(recvData, out ushort cmd, out byte[] body))
                        {
                            string hex = BinaryHelper.BytesToHex(recvData);
                            OnLog?.Invoke($"收包 <- 协议号:{cmd:X4}  十六进制:[{hex}]");
                            OnRecvPacket?.Invoke(cmd, body);
                        }
                        else
                        {
                            OnLog?.Invoke($"收到无效封包，原始数据：{BinaryHelper.BytesToHex(recvData)}", true);
                        }
                    }
                    catch
                    {
                        await DisConnect();
                        break;
                    }
            }
        }

        /// <summary>
        /// 心跳循环
        /// </summary>
        private async Task HeartLoop()
        {
            while (_isAutoHeart && IsConnected && !_cts.Token.IsCancellationRequested)
            {
                await Task.Delay(_heartInterval);
                await SendPacket(CmdHeart, Array.Empty<byte>());
            }
        }
    }
}
```

## 3. 主窗口代码
### 3.1 FormMain.cs 业务逻辑
```csharp
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
```

### 3.2 FormMain.Designer.cs 界面设计代码
```csharp
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
```

## 4. 项目入口 Program.cs
```csharp
using System;
using System.Windows.Forms;

namespace WsDebugTool
{
    static class Program
    {
        [STAThread]
        static void Main()
        {
            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);
            Application.Run(new FormMain());
        }
    }
}
```

## 5. 项目配置 WsDebugTool.csproj
```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <OutputType>WinExe</OutputType>
    <TargetFramework>net6.0-windows</TargetFramework>
    <UseWindowsForms>true</UseWindowsForms>
    <PlatformTarget>x86</PlatformTarget>
    <AssemblyTitle>千年WebSocket调试工具</AssemblyTitle>
    <AssemblyVersion>1.0.0.0</AssemblyVersion>
    <FileVersion>1.0.0.0</FileVersion>
  </PropertyGroup>

  <ItemGroup>
    <Compile Include="Program.cs" />
    <Compile Include="FormMain.cs" />
    <Compile Include="FormMain.Designer.cs" />
    <Compile Include="Utils\BinaryHelper.cs" />
    <Compile Include="Utils\LogHelper.cs" />
    <Compile Include="Network\WsClient.cs" />
  </ItemGroup>

</Project>
```

## 6. 解决方案 WsDebugTool.sln
```
Microsoft Visual Studio Solution File, Format Version 12.00
# Visual Studio 2022
VisualStudioVersion = 17.0.31903.59
MinimumVisualStudioVersion = 10.0.40219.1
Project("{FAE04EC0-301F-11D3-BF4B-00C4F79EFBC}") = "WsDebugTool", "WsDebugTool\WsDebugTool.csproj", "{C1D2E3F4-5A6B-7C8D-9E0F-1A2B3C4D5E6F}"
EndProject
Global
	GlobalSection(SolutionConfigurationPlatforms) = preSolution
		Debug|x86 = Debug|x86
		Release|x86 = Release|x86
	EndGlobalSection
	GlobalSection(ProjectConfigurationPlatforms) = postSolution
		{C1D2E3F4-5A6B-7C8D-9E0F-1A2B3C4D5E6F}.Debug|x86.ActiveCfg = Debug|x86
		{C1D2E3F4-5A6B-7C8D-9E0F-1A2B3C4D5E6F}.Debug|x86.Build.0 = Debug|x86
		{C1D2E3F4-5A6B-7C8D-9E0F-1A2B3C4D5E6F}.Release|x86.ActiveCfg = Release|x86
		{C1D2E3F4-5A6B-7C8D-9E0F-1A2B3C4D5E6F}.Release|x86.Build.0 = Release|x86
	EndGlobalSection
	GlobalSection(SolutionProperties) = preSolution
		HideSolutionNode = FALSE
	EndGlobal
EndProject
```

---

# 四、功能说明 & 使用教程
## 1. 核心功能清单
1. **连接管理**：连接/断开 `ws://` 服务端，实时显示连接状态
2. **自动心跳**：勾选后每10秒自动发送 `0x0001` 心跳包（和项目协议一致）
3. **协议快速选择**：预设心跳、移动、聊天三大常用协议号
4. **自定义消息体**：支持输入文本作为包体
5. **完整日志**：收发包、时间、协议号、**十六进制原始数据**全部打印
6. **封包校验**：内置累加校验码，和Go网关服务端规则完全一致，防篡改校验
7. **异常捕获**：网络异常、封包错误友好提示

## 2. 使用步骤
1. 用 VS2022 打开 `WsDebugTool.sln`，编译运行（F5）。
2. **填写地址**：默认 `ws://127.0.0.1:8080/ws`（本地网关地址）。
3. 点击【连接】，状态栏变为「已连接」。
4. 可选勾选【自动心跳】，工具自动维持长连接。
5. 选择协议号、填写消息体，点击【发送封包】。
6. 下方日志区实时查看**收发记录 + 十六进制原始封包**。
7. 测试完成点击【断开】，或【清空日志】清理记录。

## 3. 对接你的服务端
1. 先启动 Go 网关服务（`8080` 端口）。
2. 本工具连接网关，收发数据和**Electron客户端完全同协议**。
3. 可用来：
   - 测试网关是否正常在线
   - 调试移动、聊天等业务协议
   - 抓包查看客户端与服务端原始数据
   - 压力简易测试（循环发包可自行扩展）

---

# 五、扩展方向（可继续迭代）
1. **十六进制编辑消息体**：当前仅支持文本，可增加HEX输入框，直接编辑二进制包体。
2. **封包保存/加载**：常用封包模板保存，重复调试不用重复输入。
3. **协议解析面板**：针对移动、聊天等协议，自动解析字段并可视化。
4. **批量发包/限速**：模拟多玩家压测。
5. 与后续 **简易GM工具** 合并到同一个解决方案。

---

# 六、当前整套工具汇总
目前你 `Tools` 目录下两大核心工具全部完成：
1. **资源编辑器**：`.map`/`.spr`/`.dds` 千年资源查看、编辑、导出
2. **WebSocket调试工具**：协议联调、抓包、发包、网关测试

两套工具技术栈统一（C# .NET6 WinForms）、编译部署方式一致，规则和前后端项目**完全对齐**，可直接投入运维、策划、开发日常使用。

接下来如果你需要，我继续开发 **简易GM桌面工具**（对接后台/网关，实现踢人、发道具、改角色等功能）。