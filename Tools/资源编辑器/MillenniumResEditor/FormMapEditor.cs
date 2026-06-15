using System;
using System.Drawing;
using System.Windows.Forms;
using MillenniumResEditor.Model;
using MillenniumResEditor.Parser;

namespace MillenniumResEditor
{
    public partial class FormMapEditor : Form
    {
        private readonly MapParser _mapParser;
        private readonly string _filePath;
        private const int TileSize = 32; // 千年标准瓦片32px
        private byte _curLowTile = 0;
        private byte _curHighTile = 0;
        private byte _curAttr = 0;

        public FormMapEditor(string mapPath)
        {
            InitializeComponent();
            _filePath = mapPath;
            _mapParser = new MapParser();
            _mapParser.Load(mapPath);
            Text = $"地图编辑器 - {System.IO.Path.GetFileName(mapPath)}";

            lblMapInfo.Text = $"地图尺寸：{_mapParser.Width} x {_mapParser.Height}";
            DrawMap();
        }

        // 绘制整张地图
        private void DrawMap()
        {
            int w = _mapParser.Width * TileSize;
            int h = _mapParser.Height * TileSize;
            var bmp = new Bitmap(w, h);
            using var g = Graphics.FromImage(bmp);
            g.Clear(Color.DarkGray);

            for (int y = 0; y < _mapParser.Height; y++)
            {
                for (int x = 0; x < _mapParser.Width; x++)
                {
                    var tile = _mapParser.GetTile(x, y);
                    // 简易色块区分（正式版替换为真实瓦片图集）
                    Color drawColor = tile.Attr == 1 ? Color.Black : Color.LightGreen;
                    g.FillRectangle(new SolidBrush(drawColor),
                        x * TileSize, y * TileSize, TileSize - 1, TileSize - 1);
                }
            }
            picMap.Image = bmp;
        }

        // 鼠标绘制瓦片（笔刷）
        private void picMap_MouseDown(object sender, MouseEventArgs e)
        {
            if (e.Button != MouseButtons.Left) return;
            // 计算瓦片坐标
            int tileX = e.X / TileSize;
            int tileY = e.Y / TileSize;

            var newTile = new MapTile
            {
                Low = _curLowTile,
                High = _curHighTile,
                Attr = _curAttr
            };
            _mapParser.SetTile(tileX, tileY, newTile);
            DrawMap();
        }

        // 保存地图
        private void btnSave_Click(object sender, EventArgs e)
        {
            try
            {
                _mapParser.Save(_filePath);
                MessageBox.Show("地图保存成功！", "提示");
            }
            catch (Exception ex)
            {
                MessageBox.Show("保存失败：" + ex.Message, "错误");
            }
        }

        // 切换为阻挡属性
        private void btnSetBlock_Click(object sender, EventArgs e)
        {
            _curAttr = 1;
            lblCurAttr.Text = "当前属性：阻挡";
        }

        // 切换为通行属性
        private void btnSetWalk_Click(object sender, EventArgs e)
        {
            _curAttr = 0;
            lblCurAttr.Text = "当前属性：通行";
        }
    }
}