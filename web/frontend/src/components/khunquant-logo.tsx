import { Link } from "@tanstack/react-router"

const sizeMap = {
  sm: { img: "w-8", text: "text-md" },
  md: { img: "w-12", text: "text-xl" },
  lg: { img: "w-16", text: "text-2xl" },
  xl: { img: "w-24", text: "text-4xl" },
}

interface KhunquantLogoProps {
  size?: keyof typeof sizeMap
  theme?: 'light' | 'dark'
}

export function KhunquantLogo({ size = "md", theme = "light" }: KhunquantLogoProps) {
  const { img, text } = sizeMap[size]
  const logoSrc = theme === 'dark' ? "/khunquant_brand_white.svg" : "/khunquant_brand_dark.svg";
   const textColor = theme === 'dark' ? "#ffffff" : "#000000";
   const textColor2 = theme === 'dark' ? "#3762e1" : "#0c2d90";
  return (
    <Link to="/" className="flex items-center ">
      <img className={img} src={logoSrc} alt="Logo" />
      <span className={`${text} font-bold tracking-tight`}>
        <span style={{ color: textColor2 }}>Khun</span>
        <span style={{ color: textColor }}>Quant</span>
      </span>
    </Link>
  )
}
