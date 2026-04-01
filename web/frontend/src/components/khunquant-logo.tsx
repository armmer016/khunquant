import { Link } from "@tanstack/react-router"

const sizeMap = {
  sm: { img: "w-8", text: "text-md" },
  md: { img: "w-12", text: "text-xl" },
  lg: { img: "w-16", text: "text-2xl" },
  xl: { img: "w-24", text: "text-4xl" },
}

interface KhunquantLogoProps {
  size?: keyof typeof sizeMap
}

export function KhunquantLogo({ size = "md" }: KhunquantLogoProps) {
  const { img, text } = sizeMap[size]
  return (
    <Link to="/" className="flex items-center gap-2">
      <img className={img} src="/khunquant-logo.png" alt="Logo" />
      <span className={`${text} font-bold tracking-tight`}>
        <span style={{ color: "#3e5db9" }}>Khun</span>
        <span style={{ color: "#ffffff" }}>Quant</span>
      </span>
    </Link>
  )
}
