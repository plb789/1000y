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