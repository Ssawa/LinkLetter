// Welcome to LinkLetter! This is the application that starts a web server
// for generating and automating community newsletters.

// ==========================================================================================
// What better place to introduce the, slightly atypical, commenting structure
// of this project than in the main application's entry point? This project aims
// to encourage learning, exploration, and discussion and I think it's important
// that the code documentation reflects that. Therefore an effort will be made,
// until I grow too lazy to keep it up, to maintain a fairly verbose, if not chatty,
// amount of commenting. Hopefully by doing this we can provide a casual introduction
// to the constructs of Golang, a walkthrough of the architecture of the code, and how
// the former informed the latter. As it turns out, this also provides a good sounding
// board for the tired programmer who needs to complain and justify himself. It also
// provides this particular programmer the opportunity to hear himself talk, which he
// just loves.
//
// It's also worth pointing out at this point that all of this has actually been written
// twice. Once, where I put serious thought and effort into writing clear and professionally
// and spent a weekend documenting the code. And the second time, on the monday I
// accidentally ran "git reset --hard" on the wrong repo and stopped caring.
// =========================================================================================

// Packages in Go can be a surprisingly confusing concept if you stop and think too much about
// them, at least I know they were for me. What is a statement like "package config" actually
// saying? How does it relate to how I import that package? Does it have to be the name of the file's
// parent folder? A "package" seems to refer to multiple things all at once. Is a package the
// name I define with "package ..." or is it the path leading to that file? I see weird URLs in
// these import statements, are packages webpages? All of these questions hopefully will be answered
// in this, and following comments.
//
// I think a lot of this is a failing on the part of the Go documentation and the language itself.
// The keyword "package", I think, should have instead just have been "namespace"; then it would
// have been easy to simply say that a Go "Package" is the combination of a folder path and a namespace.
// When you say "package config" at the beginning of a file, all you are saying is that when that file
// is imported somewhere else, all of it's elements are accessible under the "config" namespace, such
// as: "config.LoadConfigs()". That's all, nothing more and nothing less.
//
// Now there are some caveats to this, the main one being that all the Go files in a folder *must*
// define the same "package". You just can't have a file with "package config" and a file with
// "package logger" in the same folder. Seems a bit arbitrary, doesn't it? I imagine if you asked the
// maintainers they would say something hand wavy about how it simplifies dependency graphing and thus
// compile times, but it still seems like a silly restriction. Regardless, that restriction is there, and
// so the habit has become that because you can only have one in a folder, the "package" should just be
// the folder's name. So as a result of all of this, we often find ourselves writing stuff like:
//
//     [package_name]/[package_name].go
//     [package_name]/[package_name]_test.go
// with each one defining:
//     package [package_name]
//
// As you can tell, that's a lot of repeating for a language that prides itself so much on its implicit
// constructs. Could this have been mitigated by allowing multiple packages in a folder or just implicitly
// determining the namespace based on the folder name, thus eliminating the package keyword entirely? Yes,
// but alas it was not.
//
// Now that's all well and good but what about *this* file? Why is this "package main"? Well giving your file
// the package of main is a bit of a special case. It just tells the compiler that this file should be
// compiled as an executable, with the "main" function the entry point, as opposed to a shared library.
package main

// Ah, Go imports. Where do we begin? As much as I hate aspects of the system, the general theory is as
// simple as can be. The first trick is to not look at those GitHub urls as URLs at all. They're nothing
// more than folder paths, relative to "$GOPATH/src", with a few additional ones (like net/http) built in
// as a part of Go's standard library. In this sense, Go's imports are not all that different from Python's
// imports or C's #include macro. So then, why *are* they GitHub URLs? Well, that's where things start to
// get ugly.
//
// You see, Go, as a language, tries to do some very cool stuff. Implicit interface implementations,
// type inferencing, first class thread support, these are all fun and interesting features that
// try to bring a C-like into the modern age. Go seems to be built around doing things that people
// thought would be "cool". So I guess, somewhere along the line, one of the developers said, "Hey,
// wouldn't it be cool if, instead of having a package manager like every other language, you just
// specified the URL of the library in your source code and it was magically imported?!" And yes,
// on a whiteboard, that's kind of cool. Unfortunately, nobody at Google took the time to think
// beyond how cool they were and realize that, implementation-wise, it sucks massive eggs.
//
// So the way it's implemented is that you "import" the url to a public repo, run "go get"
// and Go will walk through all your source code, find those URLs and download them to your
// $GOPATH/src folder. Convenient? Sure! But you might be wondering things like:
//
// Q: How do I specify the version of the library I want to use?
// A: You don't. You pull from master and that's it.
//
// Q: Doesn't that mean a library maintainer can easily break backwards compatibility and there's nothing
//    I can easily do about it?
// A: Yes.
//
// Q: Doesn't that mean a library maintainer can push a security exploit to master and it would automatically
//    get compiled into my application?
// A: Yes it does.
//
// Q: github.com/Ssawa/LinkLetter isn't actually a URL to a git repo, that leads to an HTML page. How does
//    Go know how to get an actual repo out of that?
// A: Go has to hardcode this information into the Go binary for sites like Github and BitBucket.
//
// Q: What if I don't want to use one of those sites?
// A: Why wouldn't you? This is Google. All your information and data should be centralized in one monolithic,
//    non redundant, system that everybody uses on. If you don't want that, you'll have to download every dependency
//    one by one yourself.
//
// Q: What if GitHub goes down?
// A: Did you forget again? This is Google. Websites don't go down. You're supposed to make yourself completely
//    dependent on them and trust them to always be available. Besides, who ever heard of a repo hosting service
//    going down? (https://code.google.com/)
//
// Q: What if I want to fork a project? Like github.com/Ssawa/LinkLetter to github.com/LocalProjects/LinkLetter?
// A: Well Go doesn't support relative imports, so you'll have to run a find and replace across your entire project
//    and change all your source code.
//
// Q: What if I want to submit a pull request from my fork? Won't the fact that I've changed literall every file
//    cause massive merge conflicts?
// A: Yes it will.
//
// These are all very valid questions and concerns, and the Go community would kindly like you to bugger off.
// The Go maintainers have taken the stance on these issues, and just about any other with the language, that
// by not changing anything they prove themselves to be principled, stoic, and very handsome; not, in fact,
// belligerently hubristic about a cockadookie system.
//
// They did end up, reluctantly, adding vendoring support, which at least provides the shadow of an idea of
// versioning. This project takes advantage of this by way of the grea GoDeps library.
import (
	"fmt"
	"net/http"

	"github.com/Ssawa/LinkLetter/config"
	"github.com/Ssawa/LinkLetter/database"
	"github.com/Ssawa/LinkLetter/logger"
	"github.com/Ssawa/LinkLetter/web"
)

// Here it is, the entry point into our system. It's a main function just like most other programming languages.
// Go finds it and compiles it so that it executes in its runtime.
//
// I've come across many different approaches to main functions over the years. Sometimes it will do nothing more
// than call a "Start" function in another file, sometimes it seems like the entire application is written in
// main.
//
// I've personally taken to the "moocher" approach. The main function should do what it needs to do, but should
// try to be as lazy about it as possible; borrowing, stealing, and leeching off of the hardwork of the rest of
// the library. In this way, you develop a clean, terse function that self-documents the behavior of the application,
// and lulls potential contributors into the project, unaware of the horrors that await them a few folders away.
func main() {

	// Shouldn't logging be initialized after we've gotten our configs? So that the config can determine things
	// like the log level and log location? Yeah, but we don't have any of those options anyway, so that simplifies
	// things.
	logger.InitLoggingDefault()

	logger.Debug.Printf("Determining configs...")
	conf := config.ParseForConfig()

	db := database.ConnectToDB(conf)

	// So performing all database migrations on application startup is probably not the best practice. Generally
	// web applications should take a laissez-faire approach, leaving it up the the developer to manage his or
	// her database migrations with another program. But this project is a constant balancing act between being
	// easy and straightforward to encourage contribution, and establishing good habits. In this case, for such
	// a simple scoped application, I think having it so that the application "just works" when you run it is
	// worth it. I'll happily deal with the looks of scorn from my DBA friends. We should, however, add flags
	// so that you can suppress this behavior, or run the migrations without starting the webapp, as it would
	// assist in debugging and management.
	database.DoMigrations(db)

	logger.Debug.Println("Creating server...")
	server := web.CreateServer(conf, db)

	logger.Info.Printf("Starting server...")
	http.ListenAndServe(fmt.Sprintf(":%d", conf.WebPort), logger.LogHTTPRequests(logger.Info, server.Router))
}
