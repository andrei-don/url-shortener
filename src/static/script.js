async function shortenUrl() {
    const urlInput = document.getElementById('urlInput');
    const result = document.getElementById('result');
    const shortUrl = document.getElementById('shortUrl');
    
    if (!urlInput.value) {
        alert('Please enter a URL');
        return;
    }
    
    try {
        const response = await fetch('/shorten', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ url: urlInput.value })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            shortUrl.href = data.short_url;
            shortUrl.textContent = data.short_url;
            result.classList.remove('hidden');
            resetCopyButton(); // Reset copy button when new URL is generated
        } else {
            alert('Error shortening URL: ' + (data.error || 'Unknown error'));
        }
    } catch (error) {
        alert('Error: ' + error.message);
    }
}

function copyToClipboard() {
    const shortUrl = document.getElementById('shortUrl');
    const copyButton = document.getElementById('copyButton');
    const copyIcon = document.getElementById('copyIcon');
    const copyText = document.getElementById('copyText');
    
    navigator.clipboard.writeText(shortUrl.href).then(() => {
        // Change to success state
        copyIcon.textContent = '✅';
        copyText.textContent = 'Copied!';
        copyButton.classList.add('copied');
        copyButton.classList.add('check-animation');
        
        // Reset after 2 seconds
        setTimeout(resetCopyButton, 2000);
    }).catch(() => {
        // Fallback for older browsers
        copyIcon.textContent = '❌';
        copyText.textContent = 'Failed';
        setTimeout(resetCopyButton, 2000);
    });
}

function resetCopyButton() {
    const copyButton = document.getElementById('copyButton');
    const copyIcon = document.getElementById('copyIcon');
    const copyText = document.getElementById('copyText');
    
    copyIcon.textContent = '📋';
    copyText.textContent = 'Copy';
    copyButton.classList.remove('copied', 'check-animation');
}

// Allow Enter key to submit
document.getElementById('urlInput').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        shortenUrl();
    }
});