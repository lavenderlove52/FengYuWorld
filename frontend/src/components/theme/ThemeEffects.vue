<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import type { ThemeClickEffectType, ThemeCursorEffectType, ThemeEffectConfig, ThemeEffectScope } from '@/types/theme'

const props = defineProps<{
  config: ThemeEffectConfig
  pageScope: Exclude<ThemeEffectScope, 'all'>
}>()

interface BackgroundParticle {
  x: number
  y: number
  size: number
  vx: number
  vy: number
  rotation: number
  rotationSpeed: number
  shape: number
  phase: number
}

type TransientKind = 'trail' | Exclude<ThemeCursorEffectType, 'off' | 'particle-trail'> | Exclude<ThemeClickEffectType, 'off'>

interface TransientParticle {
  x: number
  y: number
  vx: number
  vy: number
  size: number
  bornAt: number
  duration: number
  opacity: number
  color: string
  kind: TransientKind
}

interface OrbitPoint {
  x: number
  y: number
  color: string
}

const canvasRef = ref<HTMLCanvasElement | null>(null)
const MAX_TRANSIENT_PARTICLES = 120
let context: CanvasRenderingContext2D | null = null
let animationFrame = 0
let resizeFrame = 0
let mediaQuery: MediaQueryList | null = null
let reducedMotion = false
let backgroundParticles: BackgroundParticle[] = []
let transientParticles: TransientParticle[] = []
let orbitPoints: OrbitPoint[] = []
let effectCleanups: Array<() => void> = []
let lastPointerParticleAt = 0
let cursorParticleCount = 0
let pointerX = window.innerWidth / 2
let pointerY = window.innerHeight / 2
let lastContinuousFrameAt = 0
const CONTINUOUS_FRAME_MS = 1000 / 30

const scopeMatches = () => props.config.scope === 'all' || props.config.scope === props.pageScope
const hasAnyEffect = () =>
  props.config.background.type !== 'off' ||
  props.config.orbit.type !== 'off' ||
  props.config.cursor.type !== 'off' ||
  props.config.click.type !== 'off'

function randomBetween(min: number, max: number) {
  return min + Math.random() * (max - min)
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max)
}

function resizeCanvas() {
  const canvas = canvasRef.value
  if (!canvas) return
  const ratio = Math.min(window.devicePixelRatio || 1, 1.25)
  const width = Math.floor(window.innerWidth * ratio)
  const height = Math.floor(window.innerHeight * ratio)
  if (context && canvas.width === width && canvas.height === height) return
  canvas.width = width
  canvas.height = height
  canvas.style.width = `${window.innerWidth}px`
  canvas.style.height = `${window.innerHeight}px`
  context = canvas.getContext('2d')
  context?.setTransform(ratio, 0, 0, ratio, 0, 0)
}

function scheduleResizeCanvas() {
  if (resizeFrame) return
  resizeFrame = requestAnimationFrame(() => {
    resizeFrame = 0
    resizeCanvas()
  })
}

function createBackgroundParticles() {
  const effect = props.config.background
  // 星座连线每帧 O(n²)，粒子过多会明显拖慢滚动
  const maxCount = effect.type === 'constellation' ? 35 : 60
  const count = Math.floor(clamp(effect.count, 0, maxCount))
  const size = clamp(effect.size, 4, 80)
  const configuredSpeed = clamp(effect.speed, 0.1, 2)
  backgroundParticles = Array.from({ length: count }, () => {
    const speed = configuredSpeed * randomBetween(0.15, 0.45)
    let vx = randomBetween(-speed, speed)
    let vy = randomBetween(-speed, speed)
    if (effect.type === 'rain') {
      vx = randomBetween(-0.35, -0.05) * configuredSpeed
      vy = randomBetween(3, 7) * configuredSpeed
    } else if (effect.type === 'bubbles') {
      vx = randomBetween(-0.2, 0.2) * configuredSpeed
      vy = randomBetween(-0.8, -0.25) * configuredSpeed
    } else if (effect.type === 'starfield') {
      vx *= 0.15
      vy = randomBetween(0.02, 0.12) * configuredSpeed
    }
    return {
      x: Math.random() * window.innerWidth,
      y: Math.random() * window.innerHeight,
      size: size * randomBetween(0.45, 1),
      vx,
      vy,
      rotation: Math.random() * Math.PI * 2,
      rotationSpeed: randomBetween(-0.006, 0.006) * configuredSpeed,
      shape: Math.floor(Math.random() * 3),
      phase: Math.random() * Math.PI * 2,
    }
  })
}

function drawShape(
  ctx: CanvasRenderingContext2D,
  kind: number,
  x: number,
  y: number,
  size: number,
  rotation: number,
) {
  ctx.save()
  ctx.translate(x, y)
  ctx.rotate(rotation)
  ctx.beginPath()
  if (kind === 0) {
    ctx.arc(0, 0, size / 2, 0, Math.PI * 2)
  } else if (kind === 1) {
    ctx.rect(-size / 2, -size / 2, size, size)
  } else {
    ctx.moveTo(0, -size / 2)
    ctx.lineTo(size / 2, size / 2)
    ctx.lineTo(-size / 2, size / 2)
    ctx.closePath()
  }
  ctx.fill()
  ctx.restore()
}

function drawSparkle(ctx: CanvasRenderingContext2D, size: number) {
  const long = size / 2
  const short = size / 7
  ctx.beginPath()
  ctx.moveTo(0, -long)
  ctx.lineTo(short, -short)
  ctx.lineTo(long, 0)
  ctx.lineTo(short, short)
  ctx.lineTo(0, long)
  ctx.lineTo(-short, short)
  ctx.lineTo(-long, 0)
  ctx.lineTo(-short, -short)
  ctx.closePath()
}

function drawClickShape(ctx: CanvasRenderingContext2D, particle: TransientParticle) {
  const { x, y, size, kind } = particle
  ctx.save()
  ctx.translate(x, y)
  ctx.fillStyle = particle.color
  ctx.strokeStyle = particle.color
  ctx.lineWidth = Math.max(1, size / 10)
  if (kind === 'trail') {
    ctx.beginPath()
    ctx.arc(0, 0, size / 2, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()
    return
  }
  if (kind === 'comet-trail') {
    ctx.beginPath()
    ctx.moveTo(-particle.vx * size * 5, -particle.vy * size * 5)
    ctx.lineTo(0, 0)
    ctx.stroke()
    ctx.beginPath()
    ctx.arc(0, 0, size / 3, 0, Math.PI * 2)
    ctx.fill()
    ctx.restore()
    return
  }
  if (kind === 'sparkle-trail' || kind === 'sparkle') {
    drawSparkle(ctx, size)
    ctx.fill()
    ctx.restore()
    return
  }
  ctx.beginPath()
  if (kind === 'heart') {
    const unit = size / 16
    ctx.moveTo(0, size * 0.42)
    ctx.bezierCurveTo(-12 * unit, -4 * unit, -8 * unit, -12 * unit, 0, -5 * unit)
    ctx.bezierCurveTo(8 * unit, -12 * unit, 12 * unit, -4 * unit, 0, size * 0.42)
  } else if (kind === 'star') {
    for (let i = 0; i < 10; i += 1) {
      const radius = i % 2 === 0 ? size / 2 : size / 4
      const angle = -Math.PI / 2 + (i * Math.PI) / 5
      const px = Math.cos(angle) * radius
      const py = Math.sin(angle) * radius
      if (i === 0) ctx.moveTo(px, py)
      else ctx.lineTo(px, py)
    }
    ctx.closePath()
  } else if (kind === 'ring') {
    ctx.arc(0, 0, size / 2, 0, Math.PI * 2)
    ctx.stroke()
    ctx.restore()
    return
  } else if (kind === 'diamond') {
    ctx.moveTo(0, -size / 2)
    ctx.lineTo(size / 2, 0)
    ctx.lineTo(0, size / 2)
    ctx.lineTo(-size / 2, 0)
    ctx.closePath()
  } else if (kind === 'flower') {
    for (let i = 0; i < 6; i += 1) {
      const angle = (Math.PI * 2 * i) / 6
      ctx.moveTo(Math.cos(angle) * size * 0.18, Math.sin(angle) * size * 0.18)
      ctx.arc(Math.cos(angle) * size * 0.28, Math.sin(angle) * size * 0.28, size * 0.2, 0, Math.PI * 2)
    }
  } else {
    ctx.arc(0, 0, size / 2, 0, Math.PI * 2)
  }
  ctx.fill()
  ctx.restore()
}

function isCursorParticle(kind: TransientKind) {
  return kind === 'trail' || kind === 'comet-trail' || kind === 'sparkle-trail'
}

function addPointerParticle(event: PointerEvent) {
  const effect = props.config.cursor
  const now = performance.now()
  if (now - lastPointerParticleAt < 24) return
  lastPointerParticleAt = now
  if (cursorParticleCount >= clamp(effect.count, 0, 50) || transientParticles.length >= MAX_TRANSIENT_PARTICLES) return
  transientParticles.push({
    x: event.clientX,
    y: event.clientY,
    vx: randomBetween(-0.12, 0.12),
    vy: randomBetween(-0.25, -0.05),
    size: clamp(effect.size, 1, 20) * randomBetween(0.65, 1),
    bornAt: now,
    duration: clamp(effect.duration, 100, 2000),
    opacity: clamp(effect.opacity, 0.05, 1),
    color: effect.color,
    kind: effect.type === 'particle-trail'
      ? 'trail'
      : effect.type === 'comet-trail'
        ? 'comet-trail'
        : 'sparkle-trail',
  })
  cursorParticleCount += 1
  ensureAnimation()
}

function onPointerMove(event: PointerEvent) {
  pointerX = event.clientX
  pointerY = event.clientY
  if (props.config.cursor.type !== 'off') addPointerParticle(event)
  if (props.config.orbit.type !== 'off') ensureAnimation()
}

function addClickParticles(event: PointerEvent) {
  const effect = props.config.click
  if (effect.type === 'off') return
  const kind: Exclude<ThemeClickEffectType, 'off'> = effect.type
  const available = MAX_TRANSIENT_PARTICLES - transientParticles.length
  const count = Math.floor(clamp(effect.count, 0, Math.min(24, available)))
  const now = performance.now()
  for (let i = 0; i < count; i += 1) {
    const angle = (Math.PI * 2 * i) / count + randomBetween(-0.2, 0.2)
    const speed = randomBetween(0.7, 2.2)
    transientParticles.push({
      x: event.clientX,
      y: event.clientY,
      vx: Math.cos(angle) * speed,
      vy: Math.sin(angle) * speed - 1,
      size: clamp(effect.size, 4, 60) * randomBetween(0.7, 1.15),
      bornAt: now,
      duration: clamp(effect.duration, 200, 3000),
      opacity: clamp(effect.opacity, 0.05, 1),
      color: effect.color,
      kind,
    })
  }
  ensureAnimation()
}

function updateBackgroundParticle(particle: BackgroundParticle) {
  particle.x += particle.vx
  particle.y += particle.vy
  particle.rotation += particle.rotationSpeed
  if (particle.x < -particle.size) particle.x = window.innerWidth + particle.size
  if (particle.x > window.innerWidth + particle.size) particle.x = -particle.size
  if (particle.y < -particle.size) particle.y = window.innerHeight + particle.size
  if (particle.y > window.innerHeight + particle.size) particle.y = -particle.size
}

function drawBackground(ctx: CanvasRenderingContext2D, now: number) {
  const effect = props.config.background
  if (effect.type === 'off') return
  const opacity = clamp(effect.opacity, 0.05, 1)
  ctx.save()
  ctx.fillStyle = effect.color
  ctx.strokeStyle = effect.color

  if (effect.type === 'constellation') {
    const maxDistance = 130
    const maxDistanceSquared = maxDistance * maxDistance
    const maxLinksPerNode = 6
    for (let i = 0; i < backgroundParticles.length; i += 1) {
      let links = 0
      const a = backgroundParticles[i]
      for (let j = i + 1; j < backgroundParticles.length; j += 1) {
        if (links >= maxLinksPerNode) break
        const b = backgroundParticles[j]
        const dx = a.x - b.x
        const dy = a.y - b.y
        const distanceSquared = dx * dx + dy * dy
        if (distanceSquared > maxDistanceSquared) continue
        links += 1
        const distance = Math.sqrt(distanceSquared)
        ctx.globalAlpha = opacity * (1 - distance / maxDistance) * 0.45
        ctx.beginPath()
        ctx.moveTo(a.x, a.y)
        ctx.lineTo(b.x, b.y)
        ctx.stroke()
      }
    }
  }

  for (const particle of backgroundParticles) {
    updateBackgroundParticle(particle)
    const twinkle = 0.55 + Math.sin(now * 0.002 + particle.phase) * 0.35
    ctx.globalAlpha = opacity * (effect.type === 'starfield' || effect.type === 'fireflies' ? twinkle : 1)
    if (effect.type === 'floating-shapes') {
      drawShape(ctx, particle.shape, particle.x, particle.y, particle.size, particle.rotation)
    } else if (effect.type === 'bubbles') {
      ctx.lineWidth = Math.max(1, particle.size / 12)
      ctx.beginPath()
      ctx.arc(particle.x, particle.y, particle.size / 2, 0, Math.PI * 2)
      ctx.stroke()
    } else if (effect.type === 'rain') {
      ctx.lineWidth = Math.max(0.75, particle.size / 18)
      ctx.beginPath()
      ctx.moveTo(particle.x, particle.y)
      ctx.lineTo(particle.x + particle.vx * 4, particle.y + particle.size * 1.8)
      ctx.stroke()
    } else {
      if (effect.type === 'fireflies') {
        ctx.shadowColor = effect.color
        ctx.shadowBlur = particle.size
      }
      ctx.beginPath()
      ctx.arc(particle.x, particle.y, effect.type === 'constellation' ? Math.max(1.5, particle.size / 8) : Math.max(1, particle.size / 5), 0, Math.PI * 2)
      ctx.fill()
      ctx.shadowBlur = 0
    }
  }
  ctx.restore()
}

function drawOrbit(ctx: CanvasRenderingContext2D, now: number) {
  const effect = props.config.orbit
  if (effect.type === 'off') return
  const count = Math.floor(clamp(effect.count, 3, 24))
  const radius = clamp(effect.radius, 20, 120)
  const size = clamp(effect.size, 1, 16)
  const baseAngle = now * 0.001 * clamp(effect.speed, 0.1, 3)
  while (orbitPoints.length < count) orbitPoints.push({ x: 0, y: 0, color: effect.color })
  orbitPoints.length = count

  for (let i = 0; i < count; i += 1) {
    const offset = (Math.PI * 2 * i) / count
    let x = 0
    let y = 0
    if (effect.type === 'atom-orbit') {
      const ring = i % 3
      const angle = baseAngle * (ring % 2 === 0 ? 1 : -1) + offset
      const localX = Math.cos(angle) * radius
      const localY = Math.sin(angle) * radius * 0.35
      const rotation = (Math.PI * ring) / 3
      x = localX * Math.cos(rotation) - localY * Math.sin(rotation)
      y = localX * Math.sin(rotation) + localY * Math.cos(rotation)
    } else if (effect.type === 'spiral-orbit') {
      const spiralRadius = radius * (0.3 + (i / count) * 0.7)
      const angle = baseAngle * 1.35 + offset * 1.8
      x = Math.cos(angle) * spiralRadius
      y = Math.sin(angle) * spiralRadius
    } else {
      const angle = baseAngle + offset
      x = Math.cos(angle) * radius
      y = Math.sin(angle) * radius * 0.72
    }
    const point = orbitPoints[i]
    point.x = pointerX + x
    point.y = pointerY + y
    point.color = effect.type === 'rainbow-orbit' ? `hsl(${(i * 360) / count + now * 0.03} 85% 62%)` : effect.color
  }

  ctx.save()
  ctx.globalAlpha = clamp(effect.opacity, 0.05, 1)
  ctx.globalCompositeOperation = 'lighter'
  if (effect.type === 'rainbow-orbit') {
    ctx.lineWidth = 0.8
    for (let i = 0; i < orbitPoints.length; i += 1) {
      const point = orbitPoints[i]
      const next = orbitPoints[(i + 1) % orbitPoints.length]
      ctx.strokeStyle = point.color
      ctx.beginPath()
      ctx.moveTo(point.x, point.y)
      ctx.lineTo(next.x, next.y)
      ctx.stroke()
    }
  } else if (effect.type === 'atom-orbit') {
    ctx.strokeStyle = effect.color
    ctx.globalAlpha *= 0.35
    for (let i = 0; i < 3; i += 1) {
      ctx.save()
      ctx.translate(pointerX, pointerY)
      ctx.rotate((Math.PI * i) / 3)
      ctx.beginPath()
      ctx.ellipse(0, 0, radius, radius * 0.35, 0, 0, Math.PI * 2)
      ctx.stroke()
      ctx.restore()
    }
    ctx.globalAlpha = clamp(effect.opacity, 0.05, 1)
  }
  for (const point of orbitPoints) {
    ctx.fillStyle = point.color
    ctx.shadowColor = point.color
    ctx.shadowBlur = size * 2
    ctx.beginPath()
    ctx.arc(point.x, point.y, size / 2, 0, Math.PI * 2)
    ctx.fill()
  }
  ctx.restore()
}

function render(now: number) {
  animationFrame = 0
  const ctx = context
  if (!ctx) return

  const hasContinuous =
    props.config.background.type !== 'off' || props.config.orbit.type !== 'off'
  const hasTransient = transientParticles.length > 0
  if (hasContinuous && !hasTransient && now - lastContinuousFrameAt < CONTINUOUS_FRAME_MS) {
    animationFrame = requestAnimationFrame(render)
    return
  }
  if (hasContinuous) lastContinuousFrameAt = now
  ctx.clearRect(0, 0, window.innerWidth, window.innerHeight)

  drawBackground(ctx, now)
  drawOrbit(ctx, now)

  let nextParticleIndex = 0
  for (const particle of transientParticles) {
    const progress = (now - particle.bornAt) / particle.duration
    if (progress >= 1) {
      if (isCursorParticle(particle.kind)) cursorParticleCount = Math.max(0, cursorParticleCount - 1)
      continue
    }
    particle.x += particle.vx
    particle.y += particle.vy
    if (!isCursorParticle(particle.kind)) particle.vy += 0.025
    ctx.globalAlpha = particle.opacity * (1 - progress)
    drawClickShape(ctx, particle)
    transientParticles[nextParticleIndex] = particle
    nextParticleIndex += 1
  }
  transientParticles.length = nextParticleIndex

  ctx.globalAlpha = 1
  if (props.config.background.type !== 'off' || props.config.orbit.type !== 'off' || transientParticles.length > 0) {
    animationFrame = requestAnimationFrame(render)
  }
}

function ensureAnimation() {
  if (!animationFrame) animationFrame = requestAnimationFrame(render)
}

function stopEffects() {
  lastContinuousFrameAt = 0
  cancelAnimationFrame(animationFrame)
  cancelAnimationFrame(resizeFrame)
  animationFrame = 0
  resizeFrame = 0
  for (const cleanup of effectCleanups.splice(0)) cleanup()
  backgroundParticles = []
  transientParticles = []
  orbitPoints = []
  cursorParticleCount = 0
  context?.clearRect(0, 0, window.innerWidth, window.innerHeight)
}

function startEffects() {
  stopEffects()
  if (document.hidden || reducedMotion || !scopeMatches() || !hasAnyEffect()) return
  resizeCanvas()

  if (props.config.background.type !== 'off') createBackgroundParticles()
  if (props.config.cursor.type !== 'off' || props.config.orbit.type !== 'off') {
    window.addEventListener('pointermove', onPointerMove, { passive: true })
    effectCleanups.push(() => window.removeEventListener('pointermove', onPointerMove))
  }
  if (props.config.click.type !== 'off') {
    window.addEventListener('pointerdown', addClickParticles, { passive: true })
    effectCleanups.push(() => window.removeEventListener('pointerdown', addClickParticles))
  }
  window.addEventListener('resize', scheduleResizeCanvas, { passive: true })
  effectCleanups.push(() => window.removeEventListener('resize', scheduleResizeCanvas))
  if (props.config.background.type !== 'off' || props.config.orbit.type !== 'off') ensureAnimation()
}

function onMotionPreferenceChange(event: MediaQueryListEvent) {
  reducedMotion = event.matches
  startEffects()
}

function onVisibilityChange() {
  if (document.hidden) stopEffects()
  else startEffects()
}

watch(() => props.config, startEffects, { deep: true })

onMounted(() => {
  mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  reducedMotion = mediaQuery.matches
  mediaQuery.addEventListener('change', onMotionPreferenceChange)
  document.addEventListener('visibilitychange', onVisibilityChange)
  startEffects()
})

onUnmounted(() => {
  stopEffects()
  mediaQuery?.removeEventListener('change', onMotionPreferenceChange)
  document.removeEventListener('visibilitychange', onVisibilityChange)
})
</script>

<template>
  <canvas
    ref="canvasRef"
    aria-hidden="true"
    class="pointer-events-none fixed inset-0 z-[5]"
  />
</template>
