"use client";

import { useEffect } from "react";

export function BootstrapLoader() {
  useEffect(() => {
    const el = document.getElementById("bootstrap-loader");
    if (el) {
      el.classList.add("hidden");
    }
  }, []);
  return null;
}
