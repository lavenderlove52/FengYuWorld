<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import type { ThemeClickEffectType, ThemeEffectConfig, ThemeEffectScope } from '@/types/theme'

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
}

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
  kind: 'trail' | ThemeClickEffectType
}

const canvasRef = ref<HTMLCanvasElement | null>(null)
const MAX_TRANSIENT_PARTICLES = 120
let context: CanvasRenderingContext2D | null = null
let animationFrame = 0
let mediaQuery: MediaQueryList | null = null
let reducedMotion = false
let backgroundParticles: BackgroundParticle[] = []
let transientParticles: TransientParticle[] = []
let effectCleanups: Array<() => void> = []
let lastPointerParticleAt = 0

const scopeMatches = () => props.config.scope === 'all' || props.config.scope === props.pageScope
const hasAnyEffect = () =>
  props.config.background.type !== 'off' ||
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
  const ratio = Math.min(window.devicePixelRatio || 1, 2)
  canvas.width = Math.floor(window.innerWidth * ratio)
  canvas.height = Math.floor(window.innerHeight * ratio)
  canvas.style.width = `${window.innerWidth}px`
  canvas.style.height = `${window.innerHeight}px`
  context = canvas.getContext('2d')
  context?.setTransform(ratio, 0, 0, ratio, 0, 0)
}

function createBackgroundParticles() {
  const effect = props.config.background
  const count = Math.floor(clamp(effect.count, 0, 60))
  const size = clamp(effect.size, 4, 80)
  const configuredSpeed = clamp(effect.speed, 0.1, 2)
  backgroundParticles = Array.from({ length: count }, () => {
    const speed = configuredSpeed * randomBetween(0.15, 0.45)
    return {
      x: Math.random() * window.innerWidth,
      y: Math.random() * window.innerHeight,
      size: size * randomBetween(0.45, 1),
      vx: randomBetween(-speed, speed),
      vy: randomBetween(-speed, speed),
      rotation: Math.random() * Math.PI * 2,
      rotationSpeed: randomBetween(-0.006, 0.006) * configuredSpeed,
      shape: Math.floor(Math.random() * 3),
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

function drawClickShape(ctx: CanvasRenderingContext2D, particle: TransientParticle) {
  const { x, y, size, kind } = particle
  ctx.save()
  ctx.translate(x, y)
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
  } else {
    ctx.arc(0, 0, size / 2, 0, Math.PI * 2)
  }
  ctx.fill()
  ctx.restore()
}

function addPointerParticle(event: PointerEvent) {
  const effect = props.config.cursor
  const now = performance.now()
  if (now - lastPointerParticleAt < 24) return
  lastPointerParticleAt = now
  const trailCount = transientParticles.filter((particle) => particle.kind === 'trail').length
  if (trailCount >= clamp(effect.count, 0, 50) || transientParticles.length >= MAX_TRANSIENT_PARTICLES) return
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
    kind: 'trail',
  })
  ensureAnimation()
}

function addClickParticles(event: PointerEvent) {
  const effect = props.config.click
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
      kind: effect.type,
    })
  }
  ensureAnimation()
}

function render(now: number) {
  animationFrame = 0
  const ctx = context
  if (!ctx) return
  ctx.clearRect(0, 0, window.innerWidth, window.innerHeight)

  if (props.config.background.type === 'floating-shapes') {
    ctx.fillStyle = props.config.background.color
    ctx.globalAlpha = clamp(props.config.background.opacity, 0.05, 1)
    for (const particle of backgroundParticles) {
      particle.x += particle.vx
      particle.y += particle.vy
      particle.rotation += particle.rotationSpeed
      if (particle.x < -particle.size) particle.x = window.innerWidth + particle.size
      if (particle.x > window.innerWidth + particle.size) particle.x = -particle.size
      if (particle.y < -particle.size) particle.y = window.innerHeight + particle.size
      if (particle.y > window.innerHeight + particle.size) particle.y = -particle.size
      drawShape(ctx, particle.shape, particle.x, particle.y, particle.size, particle.rotation)
    }
  }

  transientParticles = transientParticles.filter((particle) => {
    const progress = (now - particle.bornAt) / particle.duration
    if (progress >= 1) return false
    particle.x += particle.vx
    particle.y += particle.vy
    if (particle.kind !== 'trail') particle.vy += 0.025
    ctx.fillStyle = particle.color
    ctx.globalAlpha = particle.opacity * (1 - progress)
    drawClickShape(ctx, particle)
    return true
  })

  ctx.globalAlpha = 1
  if (props.config.background.type === 'floating-shapes' || transientParticles.length > 0) {
    animationFrame = requestAnimationFrame(render)
  }
}

function ensureAnimation() {
  if (!animationFrame) animationFrame = requestAnimationFrame(render)
}

function stopEffects() {
  cancelAnimationFrame(animationFrame)
  animationFrame = 0
  for (const cleanup of effectCleanups.splice(0)) cleanup()
  backgroundParticles = []
  transientParticles = []
  context?.clearRect(0, 0, window.innerWidth, window.innerHeight)
}

function startEffects() {
  stopEffects()
  if (document.hidden || reducedMotion || !scopeMatches() || !hasAnyEffect()) return
  resizeCanvas()

  if (props.config.background.type === 'floating-shapes') createBackgroundParticles()
  if (props.config.cursor.type === 'particle-trail') {
    window.addEventListener('pointermove', addPointerParticle, { passive: true })
    effectCleanups.push(() => window.removeEventListener('pointermove', addPointerParticle))
  }
  if (props.config.click.type !== 'off') {
    window.addEventListener('pointerdown', addClickParticles, { passive: true })
    effectCleanups.push(() => window.removeEventListener('pointerdown', addClickParticles))
  }
  window.addEventListener('resize', resizeCanvas, { passive: true })
  effectCleanups.push(() => window.removeEventListener('resize', resizeCanvas))
  if (props.config.background.type === 'floating-shapes') ensureAnimation()
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
    class="pointer-events-none fixed inset-0 z-20"
  />
</template>
