# utils.py - Fixed handlers
def htmx_current_time(request):
    """
    Returns current server time
    """
    try:
        from datetime import datetime
        now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        return f'<span class="timestamp">Server time: {now}</span>'
    except Exception as e:
        return f'<span class="timestamp">Error: {str(e)}</span>'

def htmx_user_agent(request):
    """
    Display user agent information
    """
    try:
        # Handle both dict and form-like request objects
        if hasattr(request, 'headers'):
            user_agent = request.headers.get('user-agent', 'Unknown')
        else:
            user_agent = request.get('HTTP_USER_AGENT', 'Unknown')

        return f'''
        <div class="info-box">
            <strong>Your Browser:</strong><br>
            <code>{user_agent}</code>
        </div>
        '''
    except Exception as e:
        return f'''
        <div class="info-box">
            <strong>Error:</strong> {str(e)}
        </div>
        '''