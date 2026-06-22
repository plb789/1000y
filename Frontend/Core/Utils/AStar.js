/**
 * A* 八方向寻路算法（优化版）
 *
 * ★ 性能优化：
 *   1. 用 Map 替代 Set + 字符串键，避免字符串拼接开销
 *   2. 用二叉堆（MinHeap）替代数组线性查找最小 f 值
 *   3. 增加寻路距离上限，避免超大地图长距离寻路卡顿
 *   4. openList 用 Map 存储，O(1) 查找已存在节点
 */
const AStar = (() => {
  const dirs = [
    [-1, 0], [1, 0], [0, -1], [0, 1],     // 上下左右（直线）
    [-1, -1], [-1, 1], [1, -1], [1, 1]    // 对角线
  ];

  // 默认寻路距离上限（曼哈顿距离）
  // 超过此距离直接返回空路径，避免超大地图长距离寻路卡顿
  const DEFAULT_MAX_DISTANCE = 100;

  class Node {
    constructor(x, y) {
      this.x = x;
      this.y = y;
      this.g = 0;
      this.h = 0;
      this.f = 0;
      this.parent = null;
      // 用于 MinHeap 中快速定位节点位置（堆数组索引）
      this.heapIndex = -1;
    }
  }

  /**
   * 启发式函数（曼哈顿距离）
   * 对 8 方向寻路理论上应使用对角线距离，但曼哈顿距离更保守且性能更好
   */
  function heuristic(x1, y1, x2, y2) {
    return Math.abs(x1 - x2) + Math.abs(y1 - y2);
  }

  /**
   * 二叉堆（最小堆）
   * 用于高效获取 f 值最小的节点
   * - push: O(log n)
   * - pop:  O(log n)
   * - decreaseKey: O(log n)
   * 相比数组线性查找 O(n)，长距离寻路性能提升显著
   */
  class MinHeap {
    constructor() {
      this.heap = [];
    }

    get size() {
      return this.heap.length;
    }

    push(node) {
      node.heapIndex = this.heap.length;
      this.heap.push(node);
      this._siftUp(node.heapIndex);
    }

    pop() {
      if (this.heap.length === 0) return null;
      const top = this.heap[0];
      const last = this.heap.pop();
      if (this.heap.length > 0) {
        this.heap[0] = last;
        last.heapIndex = 0;
        this._siftDown(0);
      }
      top.heapIndex = -1;
      return top;
    }

    /**
     * 节点 f 值减小时，向上调整堆
     */
    decreaseKey(node) {
      if (node.heapIndex < 0) return;
      this._siftUp(node.heapIndex);
    }

    _siftUp(idx) {
      const item = this.heap[idx];
      while (idx > 0) {
        const parentIdx = (idx - 1) >> 1;
        const parent = this.heap[parentIdx];
        if (item.f >= parent.f) break;
        this.heap[idx] = parent;
        parent.heapIndex = idx;
        idx = parentIdx;
      }
      this.heap[idx] = item;
      item.heapIndex = idx;
    }

    _siftDown(idx) {
      const size = this.heap.length;
      const item = this.heap[idx];
      while (true) {
        const left = 2 * idx + 1;
        const right = 2 * idx + 2;
        let smallest = idx;
        let smallestNode = item;

        if (left < size && this.heap[left].f < smallestNode.f) {
          smallest = left;
          smallestNode = this.heap[left];
        }
        if (right < size && this.heap[right].f < smallestNode.f) {
          smallest = right;
          smallestNode = this.heap[right];
        }

        if (smallest === idx) break;

        // 交换
        this.heap[idx] = smallestNode;
        smallestNode.heapIndex = idx;
        this.heap[smallest] = item;
        idx = smallest;
      }
      item.heapIndex = idx;
    }
  }

  /**
   * A* 寻路主函数
   * @param {number} startX - 起点X
   * @param {number} startY - 起点Y
   * @param {number} endX - 终点X
   * @param {number} endY - 终点Y
   * @param {Uint8Array|number[]} collision - 碰撞数组（一维）
   * @param {number} mapW - 地图宽度
   * @param {number} mapH - 地图高度
   * @param {object} [options] - 可选参数
   * @param {number} [options.maxDistance=100] - 寻路距离上限（曼哈顿距离）
   * @returns {Array<{x,y}>} 路径数组，空数组表示无法到达
   */
  function findPath(startX, startY, endX, endY, collision, mapW, mapH, options = {}) {
    // 参数校验
    if (startX < 0 || startY < 0 || endX < 0 || endY < 0) return [];
    if (startX >= mapW || startY >= mapH || endX >= mapW || endY >= mapH) return [];
    if (collision[startY * mapW + startX] === 1) return [];
    if (collision[endY * mapW + endX] === 1) return [];
    if (startX === endX && startY === endY) return [{x: startX, y: startY}];

    // 寻路距离上限检查
    const maxDistance = options.maxDistance !== undefined ? options.maxDistance : DEFAULT_MAX_DISTANCE;
    const distance = heuristic(startX, startY, endX, endY);
    if (distance > maxDistance) {
      // 距离过远，直接返回空路径（避免长距离寻路卡顿）
      // 调用方可以分段寻路或提示玩家
      return [];
    }

    // ★ 优化1：openList 用 MinHeap 替代数组
    const openHeap = new MinHeap();
    // ★ 优化2：用 Map 存储 openList 中的节点，O(1) 查找
    //          键为 y * mapW + x（数字键，比字符串键更快）
    const openMap = new Map();
    // ★ 优化3：closeSet 用 Map 替代 Set，避免字符串拼接
    //          键为数字索引 y * mapW + x
    const closeMap = new Map();

    const startNode = new Node(startX, startY);
    startNode.h = heuristic(startX, startY, endX, endY);
    startNode.f = startNode.h;
    openHeap.push(startNode);
    openMap.set(startY * mapW + startX, startNode);

    // 安全阀：限制最大探索节点数，防止异常情况死循环
    const maxIterations = maxDistance * maxDistance * 4;
    let iterations = 0;

    while (openHeap.size > 0) {
      iterations++;
      if (iterations > maxIterations) break;

      // 取出 f 值最小的节点（O(log n)）
      const curr = openHeap.pop();
      const currKey = curr.y * mapW + curr.x;
      openMap.delete(currKey);
      closeMap.set(currKey, curr);

      // 到达终点
      if (curr.x === endX && curr.y === endY) {
        const path = [];
        let temp = curr;
        while (temp) {
          path.unshift({ x: temp.x, y: temp.y });
          temp = temp.parent;
        }
        return path;
      }

      // 遍历 8 个方向
      for (let i = 0; i < dirs.length; i++) {
        const dx = dirs[i][0];
        const dy = dirs[i][1];
        const nx = curr.x + dx;
        const ny = curr.y + dy;

        // 边界检查
        if (nx < 0 || ny < 0 || nx >= mapW || ny >= mapH) continue;

        // 碰撞检查
        if (collision[ny * mapW + nx] === 1) continue;

        const neighborKey = ny * mapW + nx;

        // 已在 close 列表中
        if (closeMap.has(neighborKey)) continue;

        // 对角线移动时，检查两个相邻直线格子是否可通行（防止穿墙）
        if (dx !== 0 && dy !== 0) {
          if (collision[curr.y * mapW + nx] === 1) continue;
          if (collision[ny * mapW + curr.x] === 1) continue;
        }

        // 计算移动代价（对角线代价 1.414，直线代价 1）
        const moveCost = (dx !== 0 && dy !== 0) ? 1.414 : 1;
        const g = curr.g + moveCost;

        // 检查是否已在 open 列表中
        const exist = openMap.get(neighborKey);

        if (!exist) {
          // 新节点
          const neighbor = new Node(nx, ny);
          neighbor.g = g;
          neighbor.h = heuristic(nx, ny, endX, endY);
          neighbor.f = neighbor.g + neighbor.h;
          neighbor.parent = curr;
          openHeap.push(neighbor);
          openMap.set(neighborKey, neighbor);
        } else if (g < exist.g) {
          // 已存在但新路径更短，更新节点
          exist.g = g;
          exist.f = exist.g + exist.h;
          exist.parent = curr;
          // f 值减小，向上调整堆
          openHeap.decreaseKey(exist);
        }
      }
    }

    return [];
  }

  return { findPath };
})();

window.AStar = AStar;
