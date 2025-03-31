"use client";

import { useTheme } from "nextra-theme-docs";
import React, { useEffect, useState, useMemo } from "react";
import { codeToHtml } from "shiki";

interface CodeStyleRenderProps {
  parsed: string;
  language: string;
}

/**
 * CodeStyleRender - Renders code with syntax highlighting
 * 
 * SECURITY NOTE: This component uses dangerouslySetInnerHTML and should only be used with
 * trusted inputs. The 'parsed' prop should be validated and sanitized before being passed
 * to this component to prevent XSS attacks. The 'language' prop is sanitized internally.
 */
const CodeStyleRender = ({ parsed, language }: CodeStyleRenderProps) => {
  const [html, setHtml] = useState<string>("");
  const theme = useTheme();

  const themeName = useMemo(() => {
    return theme.resolvedTheme === "dark" ? "github-dark" : "github-light";
  }, [theme.resolvedTheme]);

  useEffect(() => {
    // Sanitize language to only allow valid language identifiers
    const safeLanguage = (language || "")
      .toLowerCase()
      .replace(/[^a-z0-9-]/g, "");
    
    const asyncHighlight = async () => {
      try {
        // We rely on shiki to properly escape the code content
        const highlightedHtml = await codeToHtml(parsed || "", {
          lang: safeLanguage || "text", // Default to 'text' if language is empty after sanitization
          theme: themeName,
        });

        setHtml(highlightedHtml);
      } catch (error) {
        console.error("Error highlighting code:", error);
        // Fallback to a safe plain-text display in case of error
        setHtml(`<pre>${
          String(parsed || "")
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
        }</pre>`);
      }
    };

    asyncHighlight();
  }, [parsed, language, themeName]);

  return (
    <>
      <div dangerouslySetInnerHTML={{ __html: html }}></div>
    </>
  );
};

export default CodeStyleRender;