import { CommonModule } from '@angular/common';
import { Component, Input, NgZone, OnDestroy, OnInit, inject } from '@angular/core';
import { gsap } from 'gsap';


// A position: fixed element is positioned relative to the viewport UNLESS an
// ancestor establishes a containing block (transform, perspective, filter,
// will-change of those, or contain). When that happens, the cursor's translate
// no longer maps to viewport coordinates, so we measure and compensate for it.
const getContainingBlock = (element: HTMLElement | null): HTMLElement | null => {
  let node = element?.parentElement ?? null;
  while (node && node !== document.documentElement) {
    const style = getComputedStyle(node);
    if (
      style.transform !== 'none' ||
      style.perspective !== 'none' ||
      style.filter !== 'none' ||
      style.willChange.includes('transform') ||
      style.willChange.includes('perspective') ||
      style.willChange.includes('filter') ||
      /paint|layout|strict|content/.test(style.contain)
    ) {
      return node;
    }
    node = node.parentElement;
  }
  return null;
};

const getContainingBlockOffset = (block: HTMLElement | null): { x: number; y: number } => {
  if (!block) return { x: 0, y: 0 };
  const rect = block.getBoundingClientRect();
  return { x: rect.left + block.clientLeft, y: rect.top + block.clientTop };
};

export interface TargetCursorProps {
  targetSelector?: string;
  spinDuration?: number;
  hideDefaultCursor?: boolean;
  hoverDuration?: number;
  parallaxOn?: boolean;
  cursorColor?: string;
  cursorColorOnTarget?: string;
}

@Component({
  selector: 'app-target-cursor',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="target-cursor-wrapper" [attr.aria-hidden]="true">
      <div class="target-cursor-dot" [style.background-color]="cursorColor"></div>
      <div class="target-cursor-corner corner-tl" [style.border-color]="cursorColor"></div>
      <div class="target-cursor-corner corner-tr" [style.border-color]="cursorColor"></div>
      <div class="target-cursor-corner corner-br" [style.border-color]="cursorColor"></div>
      <div class="target-cursor-corner corner-bl" [style.border-color]="cursorColor"></div>
    </div>
  `,
  styles: [
    `
    .target-cursor-wrapper {
      position: fixed;
      top: 0;
      left: 0;
      width: 0;
      height: 0;
      pointer-events: none;
      z-index: 9999;
      mix-blend-mode: difference;
      transform: translate(-50%, -50%);
    }

    .target-cursor-dot {
      position: absolute;
      left: 50%;
      top: 50%;
      width: 4px;
      height: 4px;
      background: #fff;
      border-radius: 50%;
      transform: translate(-50%, -50%);
      will-change: transform;
    }

    .target-cursor-corner {
      position: absolute;
      left: 50%;
      top: 50%;
      width: 12px;
      height: 12px;
      border: 3px solid #fff;
      will-change: transform;
    }

    .corner-tl {
      transform: translate(-150%, -150%);
      border-right: none;
      border-bottom: none;
    }

    .corner-tr {
      transform: translate(50%, -150%);
      border-left: none;
      border-bottom: none;
    }

    .corner-br {
      transform: translate(50%, 50%);
      border-left: none;
      border-top: none;
    }

    .corner-bl {
      transform: translate(-150%, 50%);
      border-right: none;
      border-top: none;
    }
  `,
  ],
})
export class TargetCursorComponent implements OnInit, OnDestroy {
  @Input() targetSelector = '.cursor-target';
  @Input() spinDuration = 2;
  @Input() hideDefaultCursor = true;
  @Input() hoverDuration = 0.2;
  @Input() parallaxOn = true;
  @Input() cursorColor = '#ffffff';
  @Input() cursorColorOnTarget?: string;

  private zone = inject(NgZone);

  private cursorEl: HTMLDivElement | null = null;
  private dotEl: HTMLDivElement | null = null;
  private cornersEls: HTMLDivElement[] = [];
  private containingBlockEl: HTMLElement | null = null;

  private spinTl: gsap.core.Timeline | null = null;

  private isActive = false;
  private targetCornerPositions: { x: number; y: number }[] | null = null;
  private activeStrength = { current: 0 };

  private resumeTimeout: ReturnType<typeof setTimeout> | null = null;
  private currentLeaveHandler: (() => void) | null = null;
  private activeTarget: Element | null = null;

  private originalCursor = '';

  private readonly isMobile = (() => {
    if (typeof window === 'undefined') return true;
    const hasTouchScreen = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    const isSmallScreen = window.innerWidth <= 768;
    const userAgent = navigator.userAgent || (navigator as any).vendor || (window as any).opera;
    const mobileRegex = /android|webos|iphone|ipad|ipod|blackberry|iemobile|opera mini/i;
    const isMobileUserAgent = mobileRegex.test(userAgent.toLowerCase());
    return (hasTouchScreen && isSmallScreen) || isMobileUserAgent;
  })();

  private readonly constants = { borderWidth: 3, cornerSize: 12 };

  ngOnInit(): void {
    if (this.isMobile) return;

    // Run outside Angular change detection.
    this.zone.runOutsideAngular(() => {
      this.initCursor();
    });
  }

  private initCursor(): void {
    this.cursorEl = document.querySelector<HTMLDivElement>('.target-cursor-wrapper');
    this.dotEl = document.querySelector<HTMLDivElement>('.target-cursor-dot');
    if (!this.cursorEl) return;

    this.cornersEls = Array.from(
      this.cursorEl.querySelectorAll<HTMLDivElement>('.target-cursor-corner')
    );

    if (this.hideDefaultCursor) {
      this.originalCursor = document.body.style.cursor;
      document.body.style.cursor = 'none';
    }

    this.containingBlockEl = getContainingBlock(this.cursorEl);

    const moveCursor = (x: number, y: number) => {
      if (!this.cursorEl) return;
      const { x: offsetX, y: offsetY } = getContainingBlockOffset(this.containingBlockEl);
      gsap.to(this.cursorEl, {
        x: x - offsetX,
        y: y - offsetY,
        duration: 0.1,
        ease: 'power3.out',
      });
    };

    const getOffset = () => getContainingBlockOffset(this.containingBlockEl);

    const cleanupTarget = (target: Element) => {
      if (this.currentLeaveHandler) {
        target.removeEventListener('mouseleave', this.currentLeaveHandler as EventListener);
      }
      this.currentLeaveHandler = null;
    };

    const { x: initialX, y: initialY } = getOffset();
    gsap.set(this.cursorEl, {
      xPercent: -50,
      yPercent: -50,
      x: window.innerWidth / 2 - initialX,
      y: window.innerHeight / 2 - initialY,
    });

    const createSpinTimeline = () => {
      if (this.spinTl) this.spinTl.kill();
      this.spinTl = gsap.timeline({ repeat: -1 }).to(this.cursorEl!, {
        rotation: '+=360',
        duration: this.spinDuration,
        ease: 'none',
      });
    };

    createSpinTimeline();

    const tickerFn = () => {
      if (!this.targetCornerPositions || !this.cursorEl || this.cornersEls.length === 0) return;
      const strength = this.activeStrength.current;
      if (strength === 0) return;

      const cursorX = gsap.getProperty(this.cursorEl, 'x') as number;
      const cursorY = gsap.getProperty(this.cursorEl, 'y') as number;

      this.cornersEls.forEach((corner, i) => {
        const currentX = gsap.getProperty(corner, 'x') as number;
        const currentY = gsap.getProperty(corner, 'y') as number;

        const targetX = this.targetCornerPositions![i].x - cursorX;
        const targetY = this.targetCornerPositions![i].y - cursorY;

        const finalX = currentX + (targetX - currentX) * strength;
        const finalY = currentY + (targetY - currentY) * strength;

        const duration = strength >= 0.99 ? (this.parallaxOn ? 0.2 : 0) : 0.05;

        gsap.to(corner, {
          x: finalX,
          y: finalY,
          duration,
          ease: duration === 0 ? 'none' : 'power1.out',
          overwrite: 'auto',
        });
      });
    };

    const onMouseMove = (e: MouseEvent) => moveCursor(e.clientX, e.clientY);
    window.addEventListener('mousemove', onMouseMove);

    const onScroll = () => {
      if (!this.activeTarget || !this.cursorEl) return;
      const { x: offsetX, y: offsetY } = getOffset();
      const mouseX = (gsap.getProperty(this.cursorEl, 'x') as number) + offsetX;
      const mouseY = (gsap.getProperty(this.cursorEl, 'y') as number) + offsetY;
      const elementUnderMouse = document.elementFromPoint(mouseX, mouseY);

      const stillOverTarget =
        elementUnderMouse &&
        (elementUnderMouse === this.activeTarget ||
          (elementUnderMouse as Element).closest(this.targetSelector) === this.activeTarget);

      if (!stillOverTarget) {
        this.currentLeaveHandler?.();
      }
    };
    window.addEventListener('scroll', onScroll, { passive: true });

    const onMouseDown = () => {
      if (!this.dotEl || !this.cursorEl) return;
      gsap.to(this.dotEl, { scale: 0.7, duration: 0.3 });
      gsap.to(this.cursorEl, { scale: 0.9, duration: 0.2 });
    };

    const onMouseUp = () => {
      if (!this.dotEl || !this.cursorEl) return;
      gsap.to(this.dotEl, { scale: 1, duration: 0.3 });
      gsap.to(this.cursorEl, { scale: 1, duration: 0.2 });
    };

    window.addEventListener('mousedown', onMouseDown);
    window.addEventListener('mouseup', onMouseUp);

    const onMouseOver = (e: MouseEvent) => {
      const directTarget = e.target as Element;
      const allTargets: Element[] = [];
      let current: Element | null = directTarget;
      while (current && current !== document.body) {
        if ((current as any).matches?.(this.targetSelector)) {
          allTargets.push(current);
        }
        current = current.parentElement;
      }
      const target = allTargets[0] ?? null;
      if (!target || !this.cursorEl || this.cornersEls.length === 0) return;
      if (this.activeTarget === target) return;

      if (this.activeTarget) cleanupTarget(this.activeTarget);
      if (this.resumeTimeout) {
        clearTimeout(this.resumeTimeout);
        this.resumeTimeout = null;
      }

      this.activeTarget = target;

      const { borderWidth, cornerSize } = this.constants;
      const rect = target.getBoundingClientRect();
      const { x: offsetX, y: offsetY } = getOffset();
      const cursorX = gsap.getProperty(this.cursorEl, 'x') as number;
      const cursorY = gsap.getProperty(this.cursorEl, 'y') as number;

      this.cornersEls.forEach((corner) => gsap.killTweensOf(corner, 'x,y'));
      gsap.killTweensOf(this.cursorEl, 'rotation');
      this.spinTl?.pause();
      gsap.set(this.cursorEl, { rotation: 0 });

      if (this.cursorColorOnTarget) {
        gsap.to(this.cornersEls, {
          borderColor: this.cursorColorOnTarget,
          duration: 0.15,
          ease: 'power2.out',
        });
        if (this.dotEl) {
          gsap.to(this.dotEl, {
            backgroundColor: this.cursorColorOnTarget,
            duration: 0.15,
            ease: 'power2.out',
          });
        }
      }

      this.targetCornerPositions = [
        { x: rect.left - borderWidth - offsetX, y: rect.top - borderWidth - offsetY },
        {
          x: rect.right + borderWidth - cornerSize - offsetX,
          y: rect.top - borderWidth - offsetY,
        },
        {
          x: rect.right + borderWidth - cornerSize - offsetX,
          y: rect.bottom + borderWidth - cornerSize - offsetY,
        },
        {
          x: rect.left - borderWidth - offsetX,
          y: rect.bottom + borderWidth - cornerSize - offsetY,
        },
      ];

      this.isActive = true;
      gsap.ticker.add(tickerFn);

      gsap.to(this.activeStrength, {
        current: 1,
        duration: this.hoverDuration,
        ease: 'power2.out',
      });

      this.cornersEls.forEach((corner, i) => {
        gsap.to(corner, {
          x: this.targetCornerPositions![i].x - cursorX,
          y: this.targetCornerPositions![i].y - cursorY,
          duration: 0.2,
          ease: 'power2.out',
        });
      });

      this.currentLeaveHandler = () => {
        gsap.ticker.remove(tickerFn);
        this.isActive = false;
        this.targetCornerPositions = null;
        gsap.set(this.activeStrength, { current: 0, overwrite: true });
        this.activeTarget = null;

        if (this.cursorColorOnTarget && this.cornersEls.length) {
          gsap.to(this.cornersEls, {
            borderColor: this.cursorColor,
            duration: 0.15,
            ease: 'power2.out',
          });
          if (this.dotEl) {
            gsap.to(this.dotEl, {
              backgroundColor: this.cursorColor,
              duration: 0.15,
              ease: 'power2.out',
            });
          }
        }

        const tlPositions = [
          { x: -cornerSize * 1.5, y: -cornerSize * 1.5 },
          { x: cornerSize * 0.5, y: -cornerSize * 1.5 },
          { x: cornerSize * 0.5, y: cornerSize * 0.5 },
          { x: -cornerSize * 1.5, y: cornerSize * 0.5 },
        ];

        this.cornersEls.forEach((corner, idx) => gsap.killTweensOf(corner, 'x,y'));
        const tl = gsap.timeline();
        this.cornersEls.forEach((corner, index) => {
          tl.to(
            corner,
            {
              x: tlPositions[index].x,
              y: tlPositions[index].y,
              duration: 0.3,
              ease: 'power3.out',
            },
            0
          );
        });

        this.resumeTimeout = setTimeout(() => {
          if (!this.activeTarget && this.cursorEl && this.spinTl) {
            const currentRotation = gsap.getProperty(this.cursorEl, 'rotation') as number;
            const normalizedRotation = currentRotation % 360;
            this.spinTl.kill();
            this.spinTl = gsap
              .timeline({ repeat: -1 })
              .to(this.cursorEl, { rotation: '+=360', duration: this.spinDuration, ease: 'none' });

            gsap.to(this.cursorEl, {
              rotation: normalizedRotation + 360,
              duration: this.spinDuration * (1 - normalizedRotation / 360),
              ease: 'none',
              onComplete: () => {
                this.spinTl?.restart();
              },
            });
          }
          this.resumeTimeout = null;
        }, 50);

        cleanupTarget(target);
      };

      target.addEventListener('mouseleave', this.currentLeaveHandler as EventListener);
    };

    window.addEventListener('mouseover', onMouseOver as EventListener, { passive: true });

    const onResize = () => {
      this.containingBlockEl = getContainingBlock(this.cursorEl);
    };
    window.addEventListener('resize', onResize);

    // Store handlers for cleanup using closures captured above.
    (this as any)._cleanup = () => {
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('scroll', onScroll);
      window.removeEventListener('mousedown', onMouseDown);
      window.removeEventListener('mouseup', onMouseUp);
      window.removeEventListener('mouseover', onMouseOver as EventListener);
      window.removeEventListener('resize', onResize);
      if (this.activeTarget) cleanupTarget(this.activeTarget);
      this.spinTl?.kill();
      if (this.hideDefaultCursor) document.body.style.cursor = this.originalCursor;
      this.isActive = false;
      this.targetCornerPositions = null;
      this.activeStrength.current = 0;
    };
  }

  ngOnDestroy(): void {
    const cleanup = (this as any)._cleanup as undefined | (() => void);
    cleanup?.();
  }
}

