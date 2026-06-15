/**
 * A* 八方向寻路算法
 */
const AStar = (() => {
  const dirs = [
    [-1, 0], [1, 0], [0, -1], [0, 1],
    [-1, -1], [-1, 1], [1, -1], [1, 1]
  ];

  class Node {
    constructor(x, y) {
      this.x = x;
      this.y = y;
      this.g = 0;
      this.h = 0;
      this.f = 0;
      this.parent = null;
    }
  }

  function heuristic(x1, y1, x2, y2) {
    return Math.abs(x1 - x2) + Math.abs(y1 - y2);
  }

  function findPath(startX, startY, endX, endY, collision, mapW, mapH) {
    if (collision[startY * mapW + startX] === 1) return [];
    if (collision[endY * mapW + endX] === 1) return [];
    if (startX === endX && startY === endY) return [];

    const openList = [];
    const closeSet = new Set();
    const startNode = new Node(startX, startY);
    const endNode = new Node(endX, endY);
    openList.push(startNode);

    while (openList.length > 0) {
      let currIdx = 0;
      for (let i = 0; i < openList.length; i++) {
        if (openList[i].f < openList[currIdx].f) currIdx = i;
      }
      const curr = openList[currIdx];
      openList.splice(currIdx, 1);
      closeSet.add(`${curr.x},${curr.y}`);

      if (curr.x === endNode.x && curr.y === endNode.y) {
        const path = [];
        let temp = curr;
        while (temp) {
          path.unshift({ x: temp.x, y: temp.y });
          temp = temp.parent;
        }
        return path;
      }

      for (const [dx, dy] of dirs) {
        const nx = curr.x + dx;
        const ny = curr.y + dy;
        const key = `${nx},${ny}`;

        if (nx < 0 || ny < 0 || nx >= mapW || ny >= mapH) continue;
        if (collision[ny * mapW + nx] === 1) continue;
        if (closeSet.has(key)) continue;

        const neighbor = new Node(nx, ny);
        const g = curr.g + 1;
        let exist = openList.find(n => n.x === nx && n.y === ny);

        if (!exist) {
          neighbor.g = g;
          neighbor.h = heuristic(nx, ny, endNode.x, endNode.y);
          neighbor.f = neighbor.g + neighbor.h;
          neighbor.parent = curr;
          openList.push(neighbor);
        } else if (g < exist.g) {
          exist.g = g;
          exist.f = exist.g + exist.h;
          exist.parent = curr;
        }
      }
    }
    return [];
  }

  return { findPath };
})();

window.AStar = AStar;