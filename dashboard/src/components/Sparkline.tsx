export function Sparkline({
    data,
    width = 240,
    height = 48,
    color = '#FFB800',
}: {
    data: number[];
    width?: number;
    height?: number;
    color?: string;
}) {
    if (!data.length) return null;
    const max = Math.max(1, ...data);
    const step = width / Math.max(1, data.length - 1);
    const path = data
        .map((v, i) => `${i === 0 ? 'M' : 'L'} ${(i * step).toFixed(1)} ${(height - (v / max) * (height - 4) - 2).toFixed(1)}`)
        .join(' ');
    const total = data.reduce((a, b) => a + b, 0);

    return (
        <svg viewBox={`0 0 ${width} ${height}`} className="w-full" style={{ height }} preserveAspectRatio="none" role="img" aria-label={`${total} events over time`}>
            <path d={`${path} L ${width} ${height} L 0 ${height} Z`} fill={`${color}14`} />
            <path d={path} fill="none" stroke={color} strokeWidth="1.5" vectorEffect="non-scaling-stroke" />
        </svg>
    );
}
