import { Directive, Input, ElementRef, OnInit, OnDestroy, AfterViewInit, ContentChildren, QueryList, Inject, PLATFORM_ID, HostListener } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { gsap } from 'gsap';

const DEFAULT_PARTICLE_COUNT = 12;
const DEFAULT_SPOTLIGHT_RADIUS = 300;
const DEFAULT_GLOW_COLOR = '132, 0, 255';
const MOBILE_BREAKPOINT = 768;

const calculateSpotlightValues = (radius: number) => ({
  proximity: radius * 0.5,
  fadeDistance: radius * 0.75
});

const createParticleElement = (x: number, y: number, color: string = DEFAULT_GLOW_COLOR): HTMLDivElement => {
  const el = document.createElement('div');
  el.className = 'particle';
  el.style.cssText = `
    position: absolute;
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background: rgba(${color}, 1);
    box-shadow: 0 0 6px rgba(${color}, 0.6);
    pointer-events: none;
    z-index: 100;
    left: ${x}px;
    top: ${y}px;
  `;
  return el;
};

@Directive({
  selector: '[appMagicBentoCard]',
  standalone: true,
  host: {
    '[class.magic-bento-card--border-glow]': 'enableBorderGlow',
    '[class.particle-container]': 'enableStars',
    '[style.--glow-color]': 'glowColor',
    '(mouseenter)': 'onMouseEnter()',
    '(mouseleave)': 'onMouseLeave()',
    '(mousemove)': 'onMouseMove($event)',
    '(click)': 'onClick($event)'
  }
})
export class MagicBentoCardDirective implements OnInit, OnDestroy {
  @Input() enableStars = true;
  @Input() enableBorderGlow = true;
  @Input() disableAnimations = false;
  @Input() particleCount = DEFAULT_PARTICLE_COUNT;
  @Input() glowColor = DEFAULT_GLOW_COLOR;
  @Input() enableTilt = false;
  @Input() clickEffect = true;
  @Input() enableMagnetism = false; // default false so it doesn't mess up simple cards

  private isHovered = false;
  private particlesInitialized = false;
  private memoizedParticles: HTMLDivElement[] = [];
  private activeParticles: HTMLDivElement[] = [];
  private timeouts: ReturnType<typeof setTimeout>[] = [];
  private magnetismAnimation: gsap.core.Tween | null = null;
  private isBrowser: boolean;

  constructor(public el: ElementRef<HTMLElement>, @Inject(PLATFORM_ID) platformId: Object) {
    this.isBrowser = isPlatformBrowser(platformId);
  }

  ngOnInit() {}

  ngOnDestroy() {
    this.clearAllParticles();
  }

  get cardElement(): HTMLElement {
    return this.el.nativeElement;
  }

  private initializeParticles() {
    if (this.particlesInitialized || !this.cardElement) return;

    const { width, height } = this.cardElement.getBoundingClientRect();
    this.memoizedParticles = Array.from({ length: this.particleCount }, () =>
      createParticleElement(Math.random() * width, Math.random() * height, this.glowColor)
    );
    this.particlesInitialized = true;
  }

  private clearAllParticles() {
    this.timeouts.forEach(clearTimeout);
    this.timeouts = [];
    this.magnetismAnimation?.kill();

    this.activeParticles.forEach(particle => {
      gsap.to(particle, {
        scale: 0,
        opacity: 0,
        duration: 0.3,
        ease: 'back.in(1.7)',
        onComplete: () => {
          particle.parentNode?.removeChild(particle);
        }
      });
    });
    this.activeParticles = [];
  }

  private animateParticles() {
    if (!this.cardElement || !this.isHovered) return;

    if (!this.particlesInitialized) {
      this.initializeParticles();
    }

    this.memoizedParticles.forEach((particle, index) => {
      const timeoutId = setTimeout(() => {
        if (!this.isHovered || !this.cardElement) return;

        const clone = particle.cloneNode(true) as HTMLDivElement;
        this.cardElement.appendChild(clone);
        this.activeParticles.push(clone);

        gsap.fromTo(clone, { scale: 0, opacity: 0 }, { scale: 1, opacity: 1, duration: 0.3, ease: 'back.out(1.7)' });

        gsap.to(clone, {
          x: (Math.random() - 0.5) * 100,
          y: (Math.random() - 0.5) * 100,
          rotation: Math.random() * 360,
          duration: 2 + Math.random() * 2,
          ease: 'none',
          repeat: -1,
          yoyo: true
        });

        gsap.to(clone, {
          opacity: 0.3,
          duration: 1.5,
          ease: 'power2.inOut',
          repeat: -1,
          yoyo: true
        });
      }, index * 100);

      this.timeouts.push(timeoutId);
    });
  }

  onMouseEnter() {
    if (this.disableAnimations || !this.isBrowser) return;
    this.isHovered = true;
    
    if (this.enableStars) {
      this.animateParticles();
    }

    if (this.enableTilt) {
      gsap.to(this.cardElement, {
        rotateX: 5,
        rotateY: 5,
        duration: 0.3,
        ease: 'power2.out',
        transformPerspective: 1000
      });
    }
  }

  onMouseLeave() {
    if (this.disableAnimations || !this.isBrowser) return;
    this.isHovered = false;
    
    if (this.enableStars) {
      this.clearAllParticles();
    }

    if (this.enableTilt) {
      gsap.to(this.cardElement, {
        rotateX: 0,
        rotateY: 0,
        duration: 0.3,
        ease: 'power2.out'
      });
    }

    if (this.enableMagnetism) {
      gsap.to(this.cardElement, {
        x: 0,
        y: 0,
        duration: 0.3,
        ease: 'power2.out'
      });
    }
  }

  onMouseMove(e: MouseEvent) {
    if (this.disableAnimations || !this.isBrowser) return;
    if (!this.enableTilt && !this.enableMagnetism) return;

    const rect = this.cardElement.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const centerX = rect.width / 2;
    const centerY = rect.height / 2;

    if (this.enableTilt) {
      const rotateX = ((y - centerY) / centerY) * -10;
      const rotateY = ((x - centerX) / centerX) * 10;

      gsap.to(this.cardElement, {
        rotateX,
        rotateY,
        duration: 0.1,
        ease: 'power2.out',
        transformPerspective: 1000
      });
    }

    if (this.enableMagnetism) {
      const magnetX = (x - centerX) * 0.05;
      const magnetY = (y - centerY) * 0.05;

      this.magnetismAnimation = gsap.to(this.cardElement, {
        x: magnetX,
        y: magnetY,
        duration: 0.3,
        ease: 'power2.out'
      });
    }
  }

  onClick(e: MouseEvent) {
    if (!this.clickEffect || this.disableAnimations || !this.isBrowser) return;

    const rect = this.cardElement.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    const maxDistance = Math.max(
      Math.hypot(x, y),
      Math.hypot(x - rect.width, y),
      Math.hypot(x, y - rect.height),
      Math.hypot(x - rect.width, y - rect.height)
    );

    const ripple = document.createElement('div');
    ripple.style.cssText = `
      position: absolute;
      width: ${maxDistance * 2}px;
      height: ${maxDistance * 2}px;
      border-radius: 50%;
      background: radial-gradient(circle, rgba(${this.glowColor}, 0.4) 0%, rgba(${this.glowColor}, 0.2) 30%, transparent 70%);
      left: ${x - maxDistance}px;
      top: ${y - maxDistance}px;
      pointer-events: none;
      z-index: 1000;
    `;

    // ensure position relative and overflow hidden for ripple
    this.cardElement.style.position = 'relative';
    this.cardElement.style.overflow = 'hidden';
    this.cardElement.appendChild(ripple);

    gsap.fromTo(
      ripple,
      { scale: 0, opacity: 1 },
      {
        scale: 1,
        opacity: 0,
        duration: 0.8,
        ease: 'power2.out',
        onComplete: () => ripple.remove()
      }
    );
  }
}

@Directive({
  selector: '[appMagicBento]',
  standalone: true
})
export class MagicBentoDirective implements OnInit, AfterViewInit, OnDestroy {
  @Input() enableSpotlight = true;
  @Input() disableAnimations = false;
  @Input() spotlightRadius = DEFAULT_SPOTLIGHT_RADIUS;
  @Input() glowColor = DEFAULT_GLOW_COLOR;

  @ContentChildren(MagicBentoCardDirective, { descendants: true }) cardDirectives!: QueryList<MagicBentoCardDirective>;

  private spotlightElement: HTMLDivElement | null = null;
  private isMobile = false;
  private isBrowser: boolean;
  private resizeListener!: () => void;
  private mouseMoveListener!: (e: MouseEvent) => void;
  private mouseLeaveListener!: () => void;

  constructor(private el: ElementRef<HTMLElement>, @Inject(PLATFORM_ID) platformId: Object) {
    this.isBrowser = isPlatformBrowser(platformId);
  }

  ngOnInit() {
    if (this.isBrowser) {
      this.checkMobile();
      this.resizeListener = () => this.checkMobile();
      window.addEventListener('resize', this.resizeListener);
      
      // ensure host has bento-section class
      this.el.nativeElement.classList.add('bento-section');
    }
  }

  ngAfterViewInit() {
    if (!this.isBrowser || this.shouldDisableAnimations() || !this.enableSpotlight) return;

    this.spotlightElement = document.createElement('div');
    this.spotlightElement.className = 'global-spotlight';
    this.spotlightElement.style.cssText = `
      position: fixed;
      width: 800px;
      height: 800px;
      border-radius: 50%;
      pointer-events: none;
      background: radial-gradient(circle,
        rgba(${this.glowColor}, 0.15) 0%,
        rgba(${this.glowColor}, 0.08) 15%,
        rgba(${this.glowColor}, 0.04) 25%,
        rgba(${this.glowColor}, 0.02) 40%,
        rgba(${this.glowColor}, 0.01) 65%,
        transparent 70%
      );
      z-index: 200;
      opacity: 0;
      transform: translate(-50%, -50%);
      mix-blend-mode: screen;
    `;
    document.body.appendChild(this.spotlightElement);

    this.mouseMoveListener = this.onGlobalMouseMove.bind(this);
    this.mouseLeaveListener = this.onGlobalMouseLeave.bind(this);
    document.addEventListener('mousemove', this.mouseMoveListener);
    document.addEventListener('mouseleave', this.mouseLeaveListener);
  }

  ngOnDestroy() {
    if (this.isBrowser) {
      window.removeEventListener('resize', this.resizeListener);
      if (this.mouseMoveListener) {
        document.removeEventListener('mousemove', this.mouseMoveListener);
        document.removeEventListener('mouseleave', this.mouseLeaveListener);
      }
      if (this.spotlightElement && this.spotlightElement.parentNode) {
        this.spotlightElement.parentNode.removeChild(this.spotlightElement);
      }
    }
  }

  shouldDisableAnimations(): boolean {
    return this.disableAnimations || this.isMobile;
  }

  private checkMobile() {
    this.isMobile = window.innerWidth <= MOBILE_BREAKPOINT;
  }

  private onGlobalMouseMove(e: MouseEvent) {
    if (!this.spotlightElement || !this.el) return;

    const section = this.el.nativeElement;
    const rect = section.getBoundingClientRect();
    const mouseInside =
      rect && e.clientX >= rect.left && e.clientX <= rect.right && e.clientY >= rect.top && e.clientY <= rect.bottom;

    const cards = this.cardDirectives.toArray().map(c => c.cardElement);

    if (!mouseInside) {
      gsap.to(this.spotlightElement, { opacity: 0, duration: 0.3, ease: 'power2.out' });
      cards.forEach(card => {
        card.style.setProperty('--glow-intensity', '0');
      });
      return;
    }

    const { proximity, fadeDistance } = calculateSpotlightValues(this.spotlightRadius);
    let minDistance = Infinity;

    cards.forEach(cardElement => {
      const cardRect = cardElement.getBoundingClientRect();
      const centerX = cardRect.left + cardRect.width / 2;
      const centerY = cardRect.top + cardRect.height / 2;
      const distance = Math.hypot(e.clientX - centerX, e.clientY - centerY) - Math.max(cardRect.width, cardRect.height) / 2;
      const effectiveDistance = Math.max(0, distance);

      minDistance = Math.min(minDistance, effectiveDistance);

      let glowIntensity = 0;
      if (effectiveDistance <= proximity) {
        glowIntensity = 1;
      } else if (effectiveDistance <= fadeDistance) {
        glowIntensity = (fadeDistance - effectiveDistance) / (fadeDistance - proximity);
      }

      this.updateCardGlowProperties(cardElement, e.clientX, e.clientY, glowIntensity, this.spotlightRadius);
    });

    gsap.to(this.spotlightElement, {
      left: e.clientX,
      top: e.clientY,
      duration: 0.1,
      ease: 'power2.out'
    });

    const targetOpacity = minDistance <= proximity
      ? 0.8
      : minDistance <= fadeDistance
        ? ((fadeDistance - minDistance) / (fadeDistance - proximity)) * 0.8
        : 0;

    gsap.to(this.spotlightElement, {
      opacity: targetOpacity,
      duration: targetOpacity > 0 ? 0.2 : 0.5,
      ease: 'power2.out'
    });
  }

  private onGlobalMouseLeave() {
    this.cardDirectives?.forEach(c => {
      c.cardElement.style.setProperty('--glow-intensity', '0');
    });
    if (this.spotlightElement) {
      gsap.to(this.spotlightElement, { opacity: 0, duration: 0.3, ease: 'power2.out' });
    }
  }

  private updateCardGlowProperties(card: HTMLElement, mouseX: number, mouseY: number, glow: number, radius: number) {
    const rect = card.getBoundingClientRect();
    const relativeX = ((mouseX - rect.left) / rect.width) * 100;
    const relativeY = ((mouseY - rect.top) / rect.height) * 100;

    card.style.setProperty('--glow-x', `${relativeX}%`);
    card.style.setProperty('--glow-y', `${relativeY}%`);
    card.style.setProperty('--glow-intensity', glow.toString());
    card.style.setProperty('--glow-radius', `${radius}px`);
  }
}
