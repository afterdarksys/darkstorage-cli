// Cloudflare Worker for install.darkstorage.io
// Redirects to the installation script on GitHub

addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request))
})

async function handleRequest(request) {
  const url = new URL(request.url)

  // Default to shell script
  let scriptUrl = 'https://raw.githubusercontent.com/afterdarksys/darkstorage-cli/main/scripts/install.sh'

  // Check for Windows PowerShell request
  if (url.pathname === '/windows.ps1' || url.pathname === '/install.ps1') {
    // If we create a Windows PowerShell script later
    scriptUrl = 'https://raw.githubusercontent.com/afterdarksys/darkstorage-cli/main/scripts/install.ps1'
  }

  // Fetch the script from GitHub
  const response = await fetch(scriptUrl)

  // Return with appropriate headers
  return new Response(response.body, {
    status: response.status,
    headers: {
      'Content-Type': 'text/plain; charset=utf-8',
      'Cache-Control': 'public, max-age=300', // Cache for 5 minutes
      'Access-Control-Allow-Origin': '*',
      'X-Content-Type-Options': 'nosniff'
    }
  })
}
