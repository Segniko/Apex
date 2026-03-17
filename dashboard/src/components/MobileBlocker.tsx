"use client";

import React from "react";

export const MobileBlocker: React.FC = () => {
  return (
    <div className="fixed inset-0 z-[9999] flex flex-col items-center justify-center bg-[#080808] text-center md:hidden overflow-hidden">
      {/* Hazard background pattern */}
      <div className="absolute inset-0 opacity-10 hazard-pattern animate-diagonal"></div>
      
      {/* Main warning container */}
      <div className="relative industrial-card p-8 border-2 border-brand scale-animate">
        <div className="mb-6 flex justify-center">
          {/* Pulsing warning icon */}
          <div className="w-16 h-16 rounded-full border-4 border-brand flex items-center justify-center animate-pulse">
            <span className="text-brand text-4xl font-bold">!</span>
          </div>
        </div>
        
        <h1 className="text-4xl md:text-6xl font-black text-brand tracking-tighter mb-4 glitch-text">
          SIGNAL LOSS
        </h1>
        
        <div className="h-1 w-full bg-brand/20 mb-6 overflow-hidden">
          <div className="h-full bg-brand animate-loading-bar"></div>
        </div>
        
        <p className="text-xl font-mono text-white/80 max-w-xs mx-auto leading-tight italic">
          "No phones allowed"
        </p>
        
        <p className="mt-8 text-xs font-mono text-neutral uppercase tracking-widest opacity-50">
          Mobile Access Restricted // Protocol Apex-01
        </p>
      </div>
      
      {/* Decorative scanline effect */}
      <div className="absolute inset-0 pointer-events-none bg-scanline opacity-20"></div>
    </div>
  );
};
