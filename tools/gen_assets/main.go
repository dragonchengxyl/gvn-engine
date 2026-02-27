// tools/gen_assets/main.go
// 为《残调·千禧罪罚》生成全部 PNG 图像资产。
// 运行：go run ./tools/gen_assets/
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
)

const (
	bgW, bgH     = 1920, 1080
	charW, charH = 480, 860
)

func main() {
	os.MkdirAll("assets/images", 0755)

	// ── 背景 ───────────────────────────────────────────────────────────────
	save("assets/images/bg_piano_room.png", genPianoRoom())
	save("assets/images/bg_rainy_street.png", genRainyStreet())
	save("assets/images/bg_music_store.png", genMusicStore())
	save("assets/images/bg_tunnel.png", genTunnel())

	// ── 角色立绘 ───────────────────────────────────────────────────────────
	save("assets/images/lei_ye_normal.png", genLeiye(moodNormal))
	save("assets/images/lei_ye_frown.png", genLeiye(moodFrown))
	save("assets/images/lei_ye_angry.png", genLeiye(moodAngry))
	save("assets/images/lei_ye_fear.png", genLeiye(moodFear))
	save("assets/images/lei_ye_gaze.png", genLeiye(moodGaze))
}

func save(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

// ── 颜色工具 ────────────────────────────────────────────────────────────────

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	t = clamp01(t)
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func addColor(base color.RGBA, dr, dg, db int) color.RGBA {
	r := clampByte(int(base.R) + dr)
	g := clampByte(int(base.G) + dg)
	b := clampByte(int(base.B) + db)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func clampByte(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// blob 在坐标 (cx,cy) 处绘制一个柔和的径向光晕
func blob(img *image.RGBA, cx, cy, radius int, c color.RGBA, maxAlpha float64) {
	r2 := float64(radius * radius)
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}
			dx, dy := float64(x-cx), float64(y-cy)
			dist2 := dx*dx + dy*dy
			if dist2 >= r2 {
				continue
			}
			t := 1.0 - math.Sqrt(dist2)/float64(radius)
			t = t * t // 二次衰减，更柔和
			alpha := maxAlpha * t
			cur := img.RGBAAt(x, y)
			blended := color.RGBA{
				R: clampByte(int(cur.R) + int(float64(c.R-cur.R)*alpha)),
				G: clampByte(int(cur.G) + int(float64(c.G-cur.G)*alpha)),
				B: clampByte(int(cur.B) + int(float64(c.B-cur.B)*alpha)),
				A: 255,
			}
			img.SetRGBA(x, y, blended)
		}
	}
}

// vGradient 用垂直渐变填充整张图
func vGradient(img *image.RGBA, top, bot color.RGBA) {
	h := img.Bounds().Dy()
	for y := 0; y < h; y++ {
		c := lerpColor(top, bot, float64(y)/float64(h))
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// ── 背景生成 ────────────────────────────────────────────────────────────────

// genPianoRoom 昏暗琴房：深棕黑基调 + 钢琴左方的暖黄台灯光晕
func genPianoRoom() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, bgW, bgH))
	top := color.RGBA{12, 8, 5, 255}
	bot := color.RGBA{6, 4, 2, 255}
	vGradient(img, top, bot)

	// 地板木纹（横向亮条）
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 60; i++ {
		y := bgH/2 + rng.Intn(bgH/2)
		brightness := rng.Intn(12) + 3
		for x := 0; x < bgW; x++ {
			c := img.RGBAAt(x, y)
			c.R = clampByte(int(c.R) + brightness)
			c.G = clampByte(int(c.G) + brightness/2)
			img.SetRGBA(x, y, c)
		}
	}

	// 台灯主光晕（钢琴右上方暖黄）
	lampX, lampY := bgW*2/3, bgH/3
	blob(img, lampX, lampY, 340, color.RGBA{220, 160, 60, 255}, 0.75)
	blob(img, lampX, lampY, 180, color.RGBA{255, 210, 110, 255}, 0.85)
	blob(img, lampX, lampY, 80, color.RGBA{255, 240, 180, 255}, 0.95)

	// 琴键反光（水平亮带）
	for y := bgH*55/100; y < bgH*58/100; y++ {
		t := float64(y-bgH*55/100) / float64(bgH*3/100)
		bright := int((1 - math.Abs(t*2-1)) * 22)
		for x := bgW / 4; x < bgW*3/4; x++ {
			c := img.RGBAAt(x, y)
			c.R = clampByte(int(c.R) + bright)
			c.G = clampByte(int(c.G) + bright)
			c.B = clampByte(int(c.B) + bright)
			img.SetRGBA(x, y, c)
		}
	}

	// 角落深阴影
	blob(img, 0, 0, 500, color.RGBA{0, 0, 0, 255}, 0.5)
	blob(img, bgW, bgH, 600, color.RGBA{0, 0, 0, 255}, 0.4)

	return img
}

// genRainyStreet 暴雨街道：深蓝灰 + 雨丝 + 街灯光晕 + 反光积水
func genRainyStreet() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, bgW, bgH))
	top := color.RGBA{15, 20, 30, 255}
	bot := color.RGBA{20, 25, 35, 255}
	vGradient(img, top, bot)

	// 地面积水（下1/3区域稍亮）
	for y := bgH * 2 / 3; y < bgH; y++ {
		t := float64(y-bgH*2/3) / float64(bgH/3)
		for x := 0; x < bgW; x++ {
			c := img.RGBAAt(x, y)
			bright := int(t * 18)
			c.B = clampByte(int(c.B) + bright)
			img.SetRGBA(x, y, c)
		}
	}

	// 街灯光晕（分布在画面上方）
	lampPositions := [][2]int{{280, 120}, {780, 100}, {1300, 115}, {1750, 90}}
	for _, lp := range lampPositions {
		blob(img, lp[0], lp[1], 300, color.RGBA{200, 180, 100, 255}, 0.35)
		blob(img, lp[0], lp[1], 100, color.RGBA{240, 220, 150, 255}, 0.6)
		// 路面反光
		blob(img, lp[0], bgH-80, 160, color.RGBA{180, 150, 80, 255}, 0.25)
	}

	// 雨丝（斜向短线）
	rng := rand.New(rand.NewSource(99))
	for i := 0; i < 2200; i++ {
		x := rng.Intn(bgW)
		y := rng.Intn(bgH)
		length := rng.Intn(30) + 15
		alpha := rng.Float64()*0.4 + 0.2
		for k := 0; k < length; k++ {
			px, py := x+k/3, y+k
			if px >= bgW || py >= bgH {
				break
			}
			c := img.RGBAAt(px, py)
			c.R = clampByte(int(c.R) + int(alpha*80))
			c.G = clampByte(int(c.G) + int(alpha*90))
			c.B = clampByte(int(c.B) + int(alpha*120))
			img.SetRGBA(px, py, c)
		}
	}

	return img
}

// genMusicStore 十字路口：深色街道 + 右侧KFC/店铺暖光 + 左侧警灯蓝白闪烁
func genMusicStore() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, bgW, bgH))
	top := color.RGBA{10, 12, 20, 255}
	bot := color.RGBA{18, 22, 30, 255}
	vGradient(img, top, bot)

	// 地面（稍暗）
	for y := bgH * 3 / 5; y < bgH; y++ {
		for x := 0; x < bgW; x++ {
			c := img.RGBAAt(x, y)
			c.R = clampByte(int(c.R) - 3)
			c.G = clampByte(int(c.G) - 3)
			img.SetRGBA(x, y, c)
		}
	}

	// 右侧店铺暖光（KFC 橙红）
	blob(img, bgW*4/5, bgH/2, 420, color.RGBA{220, 80, 20, 255}, 0.45)
	blob(img, bgW*4/5, bgH/2, 200, color.RGBA{255, 140, 60, 255}, 0.55)
	// 店铺橱窗矩形亮块
	for y := bgH * 2 / 5; y < bgH*3/5; y++ {
		for x := bgW * 7 / 10; x < bgW*9/10; x++ {
			t := 0.6 - math.Abs(float64(x-bgW*8/10)/float64(bgW/5))
			if t < 0 {
				t = 0
			}
			c := img.RGBAAt(x, y)
			c.R = clampByte(int(c.R) + int(t*120))
			c.G = clampByte(int(c.G) + int(t*50))
			img.SetRGBA(x, y, c)
		}
	}

	// 左侧警灯（蓝白光晕）
	blob(img, bgW/8, bgH/4, 200, color.RGBA{60, 100, 220, 255}, 0.45)
	blob(img, bgW/8, bgH/4, 80, color.RGBA{180, 200, 255, 255}, 0.65)

	// 雨丝（同街道但更少）
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 1400; i++ {
		x := rng.Intn(bgW)
		y := rng.Intn(bgH)
		length := rng.Intn(20) + 10
		for k := 0; k < length; k++ {
			px, py := x+k/3, y+k
			if px >= bgW || py >= bgH {
				break
			}
			c := img.RGBAAt(px, py)
			c.B = clampByte(int(c.B) + 35)
			c.R = clampByte(int(c.R) + 20)
			c.G = clampByte(int(c.G) + 25)
			img.SetRGBA(px, py, c)
		}
	}

	return img
}

// genTunnel 地下通道入口：近乎漆黑 + 深处极微弱的灰光
func genTunnel() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, bgW, bgH))
	top := color.RGBA{8, 8, 12, 255}
	bot := color.RGBA{3, 3, 5, 255}
	vGradient(img, top, bot)

	// 通道入口拱形（深灰轮廓）
	cx, cy := bgW/2, bgH*3/5
	for y := 0; y < bgH; y++ {
		for x := 0; x < bgW; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			// 椭圆：宽300，高400
			ellipse := (dx*dx)/(300*300) + (dy*dy)/(400*400)
			if ellipse < 1.0 && y < cy {
				c := img.RGBAAt(x, y)
				dim := clamp01(1.0 - ellipse)
				c.R = clampByte(int(c.R) + int(dim*18))
				c.G = clampByte(int(c.G) + int(dim*18))
				c.B = clampByte(int(c.B) + int(dim*22))
				img.SetRGBA(x, y, c)
			}
		}
	}

	// 极微弱的远处出口光（通道深处）
	blob(img, bgW/2, bgH/3, 80, color.RGBA{50, 50, 60, 255}, 0.4)
	blob(img, bgW/2, bgH/3, 30, color.RGBA{80, 80, 100, 255}, 0.5)

	// 台阶边缘（水平亮线）
	for i := 0; i < 5; i++ {
		y := bgH*3/5 + i*40
		if y >= bgH {
			break
		}
		for x := bgW/2 - 280; x <= bgW/2+280; x++ {
			if x < 0 || x >= bgW {
				continue
			}
			c := img.RGBAAt(x, y)
			c.R = clampByte(int(c.R) + 15)
			c.G = clampByte(int(c.G) + 15)
			c.B = clampByte(int(c.B) + 20)
			img.SetRGBA(x, y, c)
		}
	}

	// 四角加深（vignette）
	for y := 0; y < bgH; y++ {
		for x := 0; x < bgW; x++ {
			dx := float64(x)/float64(bgW)*2 - 1
			dy := float64(y)/float64(bgH)*2 - 1
			vignette := math.Sqrt(dx*dx+dy*dy) / math.Sqrt2
			dark := int(vignette * vignette * 25)
			c := img.RGBAAt(x, y)
			c.R = clampByte(int(c.R) - dark)
			c.G = clampByte(int(c.G) - dark)
			c.B = clampByte(int(c.B) - dark)
			img.SetRGBA(x, y, c)
		}
	}

	return img
}

// ── 角色立绘 ────────────────────────────────────────────────────────────────

type mood int

const (
	moodNormal mood = iota
	moodFrown
	moodAngry
	moodFear
	moodGaze
)

// genLeiye 生成雷业的立绘（简化型剪影 + 色彩情绪表达）
func genLeiye(m mood) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, charW, charH))

	// 基础颜色：东北虎（橙棕）
	bodyBase := color.RGBA{180, 100, 30, 255}
	stripeCol := color.RGBA{80, 40, 10, 255}
	// 西装颜色
	suitCol := color.RGBA{40, 40, 50, 255}
	// 情绪叠加
	var moodTint color.RGBA
	var moodStrength float64
	switch m {
	case moodNormal:
		moodTint = color.RGBA{0, 0, 0, 255}
		moodStrength = 0
	case moodFrown:
		moodTint = color.RGBA{60, 80, 140, 255} // 蓝（寒冷/淋湿）
		moodStrength = 0.22
	case moodAngry:
		moodTint = color.RGBA{200, 40, 20, 255} // 红（愤怒）
		moodStrength = 0.25
	case moodFear:
		moodTint = color.RGBA{160, 160, 175, 255} // 灰（恐惧/苍白）
		moodStrength = 0.32
	case moodGaze:
		moodTint = color.RGBA{10, 10, 30, 255} // 近黑（凝视深渊）
		moodStrength = 0.40
	}

	applyTint := func(c color.RGBA) color.RGBA {
		return color.RGBA{
			R: clampByte(int(float64(c.R)*(1-moodStrength) + float64(moodTint.R)*moodStrength)),
			G: clampByte(int(float64(c.G)*(1-moodStrength) + float64(moodTint.G)*moodStrength)),
			B: clampByte(int(float64(c.B)*(1-moodStrength) + float64(moodTint.B)*moodStrength)),
			A: c.A,
		}
	}

	// ── 身体：西装剪影 ─────────────────────────────────────────────────────
	// 躯干（从 y=350 到 y=860，宽度 200-350）
	fillTrapezoid(img, charW/2, 350, charH, 160, 210, applyTint(suitCol))

	// 高领毛衣（灰色，颈部区域）
	fillRect(img, charW/2-45, 290, 90, 80, applyTint(color.RGBA{80, 80, 82, 255}))

	// 西装翻领（深色三角）
	fillTri(img, charW/2-30, 340, charW/2+30, 340, charW/2-80, 460, applyTint(color.RGBA{25, 25, 32, 255}))
	fillTri(img, charW/2+30, 340, charW/2-30, 340, charW/2+80, 460, applyTint(color.RGBA{25, 25, 32, 255}))

	// ── 头部 ────────────────────────────────────────────────────────────────
	headCX, headCY, headR := charW/2, 210, 125
	fillCircle(img, headCX, headCY, headR, applyTint(bodyBase))

	// 老虎耳朵（三角形）
	earCol := applyTint(bodyBase)
	innerEar := applyTint(color.RGBA{220, 130, 80, 255})
	// 左耳
	fillTri(img, headCX-90, headCY-70, headCX-50, headCY-130, headCX-30, headCY-80, earCol)
	fillTri(img, headCX-80, headCY-75, headCX-55, headCY-115, headCX-40, headCY-82, innerEar)
	// 右耳
	fillTri(img, headCX+90, headCY-70, headCX+50, headCY-130, headCX+30, headCY-80, earCol)
	fillTri(img, headCX+80, headCY-75, headCX+55, headCY-115, headCX+40, headCY-82, innerEar)

	// 面部（稍浅区域）
	fillCircle(img, headCX, headCY+20, 75, applyTint(color.RGBA{210, 140, 70, 255}))

	// 虎纹（深色条纹）
	drawStripe(img, headCX-60, headCY-40, 20, 50, -15, applyTint(stripeCol))
	drawStripe(img, headCX+40, headCY-40, 20, 50, 15, applyTint(stripeCol))
	drawStripe(img, headCX-30, headCY-80, 15, 40, 5, applyTint(stripeCol))
	drawStripe(img, headCX+15, headCY-80, 15, 40, -5, applyTint(stripeCol))

	// 眼睛（根据情绪变化）
	eyeCol := applyTint(color.RGBA{255, 200, 50, 255}) // 虎眼：金黄
	pupilCol := applyTint(color.RGBA{10, 5, 0, 255})
	switch m {
	case moodAngry:
		// 眯眼（怒）
		drawEyeAngry(img, headCX-38, headCY+5, eyeCol, pupilCol)
		drawEyeAngry(img, headCX+38, headCY+5, eyeCol, pupilCol)
	case moodFear:
		// 瞪大眼（恐惧）
		fillCircle(img, headCX-38, headCY+5, 16, eyeCol)
		fillCircle(img, headCX+38, headCY+5, 16, eyeCol)
		fillCircle(img, headCX-38, headCY+5, 10, pupilCol)
		fillCircle(img, headCX+38, headCY+5, 10, pupilCol)
	default:
		fillCircle(img, headCX-38, headCY+5, 13, eyeCol)
		fillCircle(img, headCX+38, headCY+5, 13, eyeCol)
		fillCircle(img, headCX-38, headCY+5, 8, pupilCol)
		fillCircle(img, headCX+38, headCY+5, 8, pupilCol)
	}

	// 鼻子
	fillCircle(img, headCX, headCY+22, 10, applyTint(color.RGBA{140, 60, 40, 255}))
	// 虎须（短线）
	drawWhisker(img, headCX, headCY+22, applyTint(color.RGBA{230, 200, 160, 255}))

	// 冷汗/雨水（moodFear / moodFrown）
	if m == moodFear || m == moodFrown {
		rng := rand.New(rand.NewSource(int64(m)))
		for i := 0; i < 8; i++ {
			sx := headCX - 80 + rng.Intn(160)
			sy := headCY - 60 + rng.Intn(140)
			dropCol := applyTint(color.RGBA{160, 185, 220, 180})
			for k := 0; k < 18; k++ {
				py := sy + k
				if py >= charH {
					break
				}
				if img.RGBAAt(sx, py).A == 0 {
					break
				}
				img.SetRGBA(sx, py, dropCol)
			}
		}
	}

	return img
}

// ── 绘图辅助函数 ────────────────────────────────────────────────────────────

func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.SetRGBA(px, py, c)
			}
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for py := cy - r; py <= cy+r; py++ {
		for px := cx - r; px <= cx+r; px++ {
			if px < 0 || px >= img.Bounds().Dx() || py < 0 || py >= img.Bounds().Dy() {
				continue
			}
			dx, dy := float64(px-cx), float64(py-cy)
			if dx*dx+dy*dy <= float64(r*r) {
				img.SetRGBA(px, py, c)
			}
		}
	}
}

// fillTrapezoid 绘制一个上窄下宽的梯形（用于身体轮廓）
func fillTrapezoid(img *image.RGBA, cx, yTop, yBot, halfTop, halfBot int, c color.RGBA) {
	for y := yTop; y <= yBot; y++ {
		if y >= img.Bounds().Dy() {
			break
		}
		t := float64(y-yTop) / float64(yBot-yTop)
		half := int(float64(halfTop) + float64(halfBot-halfTop)*t)
		for x := cx - half; x <= cx+half; x++ {
			if x >= 0 && x < img.Bounds().Dx() {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func fillTri(img *image.RGBA, x1, y1, x2, y2, x3, y3 int, c color.RGBA) {
	minY := min3(y1, y2, y3)
	maxY := max3(y1, y2, y3)
	for y := minY; y <= maxY; y++ {
		xs := scanlineTri(y, x1, y1, x2, y2, x3, y3)
		if len(xs) < 2 {
			continue
		}
		for x := xs[0]; x <= xs[1]; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func scanlineTri(y, x1, y1, x2, y2, x3, y3 int) []int {
	var xs []int
	edges := [][4]int{{x1, y1, x2, y2}, {x2, y2, x3, y3}, {x3, y3, x1, y1}}
	for _, e := range edges {
		ax, ay, bx, by := e[0], e[1], e[2], e[3]
		if (ay <= y && y < by) || (by <= y && y < ay) {
			x := ax + (y-ay)*(bx-ax)/(by-ay)
			xs = append(xs, x)
		}
	}
	if len(xs) < 2 {
		return xs
	}
	if xs[0] > xs[1] {
		xs[0], xs[1] = xs[1], xs[0]
	}
	return xs
}

func drawStripe(img *image.RGBA, x, y, w, h, angle int, c color.RGBA) {
	for py := y; py < y+h; py++ {
		offset := int(float64(py-y) * math.Tan(float64(angle)*math.Pi/180))
		for px := x + offset; px < x+w+offset; px++ {
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.SetRGBA(px, py, c)
			}
		}
	}
}

func drawEyeAngry(img *image.RGBA, cx, cy int, eyeCol, pupilCol color.RGBA) {
	// 眯眼：扁椭圆（愤怒的半闭眼）
	for py := cy - 6; py <= cy+6; py++ {
		halfW := int(math.Sqrt(math.Max(0, 1-float64((py-cy)*(py-cy))/36.0)) * 14)
		for px := cx - halfW; px <= cx+halfW; px++ {
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.SetRGBA(px, py, eyeCol)
			}
		}
	}
	// 竖瞳孔
	for py := cy - 8; py <= cy+8; py++ {
		for px := cx - 3; px <= cx+3; px++ {
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.SetRGBA(px, py, pupilCol)
			}
		}
	}
}

func drawWhisker(img *image.RGBA, cx, cy int, c color.RGBA) {
	for i := -3; i <= 3; i++ {
		py := cy + i/2
		// 左须
		for x := cx - 60; x < cx-15; x++ {
			if x >= 0 && py >= 0 && py < img.Bounds().Dy() {
				img.SetRGBA(x, py, c)
			}
		}
		// 右须
		for x := cx + 15; x < cx+60; x++ {
			if x < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.SetRGBA(x, py, c)
			}
		}
	}
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max3(a, b, c int) int {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}
