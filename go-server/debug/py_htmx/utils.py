# Utility functions for HTMLnoJS demo

def htmx_current_time(request):
    """
    Returns current server time
    """
    from datetime import datetime
    now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    return f'<span class="timestamp">Server time: {now}</span>'

def htmx_user_agent(request):
    """
    Display user agent information
    """
    user_agent = request.get('HTTP_USER_AGENT', 'Unknown')
    return f'''
    <div class="info-box">
        <strong>Your Browser:</strong><br>
        <code>{user_agent}</code>
    </div>
    '''