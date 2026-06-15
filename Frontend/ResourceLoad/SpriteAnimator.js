/**
 * SPR 动画播放控制器
 */
class SpriteAnimator {
  constructor(sprParser) {
    this.parser = sprParser;
    this.curFrame = 0;
    this.frameRate = 10;
    this.frameTimer = 0;
    this.isLoop = true;
    this.actionStart = 0;
    this.actionEnd = 0;
  }

  setAction(start, end, loop = true) {
    this.actionStart = start;
    this.actionEnd = end;
    this.curFrame = start;
    this.isLoop = loop;
    this.frameTimer = 0;
  }

  update(deltaTime = 16) {
    this.frameTimer += deltaTime;
    const interval = 1000 / this.frameRate;
    if (this.frameTimer >= interval) {
      this.frameTimer = 0;
      this.curFrame++;
      if (this.curFrame > this.actionEnd) {
        this.curFrame = this.isLoop ? this.actionStart : this.actionEnd;
      }
    }
  }

  draw(ctx, x, y) {
    const frame = this.parser.getFrame(this.curFrame);
    if (!frame) return;
    ctx.putImageData(frame.imgData, x - frame.offX, y - frame.offY);
  }
}

window.SpriteAnimator = SpriteAnimator;