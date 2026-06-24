import { MobileBlocker } from "@/components/MobileBlocker";
import { TacticalChat } from "@/components/TacticalChat";

// The live HUD is a dense, desktop-first surface. The mobile gate and the
// floating forensics chat are scoped here so the marketing site and docs
// stay fully accessible on every device.
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <MobileBlocker />
      {children}
      <TacticalChat />
    </>
  );
}
