<!-- install gorm dan driver mysql -->

go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql

<!-- air untuk reload go -->

go install github.com/air-verse/air@v1.40.0

go install github.com/air-verse/air@latest <!--go versi terbaru 1.25 -->

go get github.com/gin-contrib/cors <!-- untuk bisa diakses lainnya -->

go get github.com/cloudinary/cloudinary-go/v2 <!-- untuk install cloudinary -->

<!-- Tambahkan ini di next.config.ts supaya akses ke url cloudinary bisa -->

images: {
remotePatterns: [
{
protocol: "https",
hostname: "res.cloudinary.com",
port: "",
pathname: "/**",
},
],
dangerouslyAllowSVG: true,
contentDispositionType: "attachment",
contentSecurityPolicy: "default-src 'self'; script-src 'none'; sandbox;",
formats: ["image/avif", "image/webp"],
},
