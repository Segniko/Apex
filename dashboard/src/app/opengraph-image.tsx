import { ImageResponse } from "next/og";

export const alt = "Apex — The Architecture of Recovery";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

// Generated social card. Renders the tactical brand as a real PNG so links
// shared on X / Slack / Discord / HN get a proper preview.
export default function OpengraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          height: "100%",
          width: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          backgroundColor: "#080808",
          padding: "72px",
          fontFamily: "sans-serif",
        }}
      >
        {/* top hazard bar */}
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            height: 10,
            backgroundImage:
              "repeating-linear-gradient(45deg, #FFB800, #FFB800 14px, #080808 14px, #080808 28px)",
          }}
        />
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <div style={{ width: 22, height: 22, backgroundColor: "#FFB800" }} />
          <div
            style={{
              color: "white",
              fontSize: 34,
              fontWeight: 900,
              fontStyle: "italic",
              letterSpacing: -1,
            }}
          >
            APEX
          </div>
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          <div
            style={{
              color: "white",
              fontSize: 92,
              fontWeight: 900,
              fontStyle: "italic",
              lineHeight: 1,
              letterSpacing: -3,
              textTransform: "uppercase",
            }}
          >
            Architecture of
          </div>
          <div
            style={{
              color: "#FFB800",
              fontSize: 92,
              fontWeight: 900,
              fontStyle: "italic",
              lineHeight: 1,
              letterSpacing: -3,
              textTransform: "uppercase",
            }}
          >
            Recovery.
          </div>
        </div>

        <div
          style={{
            color: "#9a9a9a",
            fontSize: 30,
            display: "flex",
            justifyContent: "space-between",
          }}
        >
          <span>Industrial-grade crash forensics. AI root-cause analysis.</span>
          <span style={{ color: "#FFB800", fontWeight: 700 }}>100% Free · OSS</span>
        </div>
      </div>
    ),
    { ...size }
  );
}
