// Package next provides the Zeichenwerk terminal UI toolkit.
//
// Zeichenwerk (german for "character works) is a complete Terminal UI toolkit
// for building interactive terminal applications. It is designed to be easy
// to use and understand.
//
// # Overview
//
// The main feature is a fluent API for building the UI in a declarative
// way using the Builder type. Even complex UIs are built easily with the
// Builder
//
// Example:
//
// ui := NewBuilder(TokyoNightTheme).
//
//	   Flex("main", false, "stretch", 0).
//	   Flex("header", true, "stretch", 2).
//	   Static("header-logo", "Zeichenwerk").
//		 Static("header-title", "Demo Application").
//		 End().
//		 Grid("body", 1, 2, true).Columns(30, -1).
//		 Cell(0, 0, 1, 1).
//		 List("navigation", "First", "Second", "Third").
//		 Cell(1, 0, 1, 1).With(content). // builder function for the content
//		 End().
//		 Flex("footer", true, "stretch", 1).
//		 Static("footer-text", "Footer").
//		 End().
//		 Run()
//
// # widgets
//
// Nearly everything including the root UI
// implements the Widget interface, which is implemented completely by the
// Component type. Creating new UI widgets is easy if the Component type
// is embedded. For simple UI widgets look at the Static or Button widgets.
package zeichenwerk
