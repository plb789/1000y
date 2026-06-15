namespace MillenniumResEditor.Model
{
    /// <summary>
    /// 千年地图单个瓦片
    /// </summary>
    public class MapTile
    {
        /// <summary>底层瓦片索引</summary>
        public byte Low { get; set; }
        /// <summary>高层瓦片索引</summary>
        public byte High { get; set; }
        /// <summary>属性 0=通行 1=阻挡 2=事件区</summary>
        public byte Attr { get; set; }
    }
}