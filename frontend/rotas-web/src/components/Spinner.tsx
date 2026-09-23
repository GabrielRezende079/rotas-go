interface SpinnerProps {
  size?: number
  className?: string
}

export function Spinner({ size = 16, className }: SpinnerProps) {
  return (
    <span
      className={`spinner${className !== undefined ? ` ${className}` : ''}`}
      style={{ width: size, height: size }}
      aria-hidden="true"
    />
  )
}