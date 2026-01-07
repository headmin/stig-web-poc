#!/usr/bin/env python3
"""
Fetch STIG data from stigviewer.com

The stigviewer.com site uses a Next.js frontend that loads data via RSC.
We can extract the STIG data by fetching the page and parsing the embedded JSON.

Alternative: Download from DISA cyber.mil (official source, but requires manual download)
"""

import json
import re
import sys
from pathlib import Path

try:
    import httpx
except ImportError:
    print("Install httpx: uv pip install httpx")
    sys.exit(1)


def fetch_stig_from_stigviewer(slug: str) -> dict | None:
    """
    Attempt to fetch STIG data from stigviewer.com
    
    The site embeds data in Next.js RSC format within script tags.
    """
    url = f"https://www.stigviewer.com/stig/{slug}"
    
    try:
        client = httpx.Client(follow_redirects=True, timeout=30)
        resp = client.get(url)
        
        if resp.status_code != 200:
            print(f"Failed to fetch {url}: HTTP {resp.status_code}")
            return None
        
        # Look for embedded JSON data in Next.js RSC format
        # Pattern: self.__next_f.push([1,"...JSON..."])
        html = resp.text
        
        # Try to find benchmark data in the page
        # This is a heuristic - the site structure may change
        json_pattern = r'"benchmarkId"\s*:\s*"([^"]+)"'
        match = re.search(json_pattern, html)
        
        if match:
            print(f"Found benchmarkId: {match.group(1)}")
            # Would need to parse the full RSC payload to extract data
            # This is complex and fragile
            
        return None
        
    except Exception as e:
        print(f"Error fetching STIG: {e}")
        return None


def parse_disa_xccdf(xccdf_path: Path) -> dict:
    """
    Parse DISA STIG XCCDF XML file into our JSON format.
    
    DISA publishes STIGs as ZIP files containing:
    - *_Manual-xccdf.xml (the STIG rules in XCCDF format)
    - *_Overview.pdf
    - *_STIG.pdf
    
    Download from: https://public.cyber.mil/stigs/downloads/
    Search for "Windows 11" and download the ZIP.
    """
    import xml.etree.ElementTree as ET
    
    # XCCDF namespace
    ns = {
        'xccdf': 'http://checklists.nist.gov/xccdf/1.1',
        'dc': 'http://purl.org/dc/elements/1.1/',
    }
    
    tree = ET.parse(xccdf_path)
    root = tree.getroot()
    
    # Extract benchmark metadata
    title = root.find('.//xccdf:title', ns)
    version = root.find('.//xccdf:version', ns)
    
    benchmark = {
        'id': root.get('id', ''),
        'benchmarkId': root.get('id', '').replace('xccdf_mil.disa.stig_benchmark_', ''),
        'slug': '',  # Generate from title
        'title': title.text if title is not None else '',
        'version': version.text if version is not None else '',
        'groups': []
    }
    
    # Generate slug from title
    if benchmark['title']:
        benchmark['slug'] = benchmark['title'].lower().replace(' ', '-').replace('_', '-')
    
    # Extract rules (Groups in XCCDF)
    for group in root.findall('.//xccdf:Group', ns):
        group_id = group.get('id', '')
        rule = group.find('xccdf:Rule', ns)
        
        if rule is None:
            continue
            
        rule_data = {
            'groupId': group_id.replace('xccdf_mil.disa.stig_group_', ''),
            'title': '',
            'ruleId': rule.get('id', ''),
            'ruleSeverity': rule.get('severity', 'medium'),
            'ruleWeight': rule.get('weight', '10.0'),
            'ruleTitle': '',
            'ruleVulnDiscussion': '',
            'ruleCheckContent': '',
            'ruleFixText': '',
            'ruleIdent': '',
        }
        
        # Extract rule details
        title_elem = rule.find('xccdf:title', ns)
        if title_elem is not None:
            rule_data['ruleTitle'] = title_elem.text or ''
            
        # Version becomes ruleVersion (e.g., WN11-00-000001)
        version_elem = rule.find('xccdf:version', ns)
        if version_elem is not None:
            rule_data['ruleVersion'] = version_elem.text or ''
            
        # Description contains vuln discussion
        desc = rule.find('xccdf:description', ns)
        if desc is not None and desc.text:
            # Parse the embedded XML in description
            try:
                desc_xml = ET.fromstring(f"<root>{desc.text}</root>")
                vuln = desc_xml.find('.//VulnDiscussion')
                if vuln is not None:
                    rule_data['ruleVulnDiscussion'] = vuln.text or ''
            except:
                rule_data['ruleVulnDiscussion'] = desc.text
                
        # Check content
        check = rule.find('.//xccdf:check-content', ns)
        if check is not None:
            rule_data['ruleCheckContent'] = check.text or ''
            
        # Fix text
        fix = rule.find('xccdf:fixtext', ns)
        if fix is not None:
            rule_data['ruleFixText'] = fix.text or ''
            
        # CCI identifier
        ident = rule.find('xccdf:ident', ns)
        if ident is not None:
            rule_data['ruleIdent'] = ident.text or ''
            
        benchmark['groups'].append(rule_data)
    
    return benchmark


def main():
    import argparse
    
    parser = argparse.ArgumentParser(description='Fetch or parse STIG data')
    parser.add_argument('--xccdf', type=Path, help='Parse DISA XCCDF XML file')
    parser.add_argument('--slug', type=str, help='Fetch from stigviewer.com by slug')
    parser.add_argument('--output', '-o', type=Path, default=Path('stig.json'), help='Output JSON file')
    
    args = parser.parse_args()
    
    if args.xccdf:
        print(f"Parsing XCCDF: {args.xccdf}")
        data = parse_disa_xccdf(args.xccdf)
        print(f"Found {len(data['groups'])} rules")
        
        with open(args.output, 'w') as f:
            json.dump(data, f, indent=2)
        print(f"Wrote: {args.output}")
        
    elif args.slug:
        print(f"Fetching from stigviewer.com: {args.slug}")
        data = fetch_stig_from_stigviewer(args.slug)
        
        if data:
            with open(args.output, 'w') as f:
                json.dump(data, f, indent=2)
            print(f"Wrote: {args.output}")
        else:
            print("Failed to fetch STIG data")
            print("\nAlternative: Download from DISA manually:")
            print("1. Go to https://public.cyber.mil/stigs/downloads/")
            print("2. Search for 'Windows 11'")
            print("3. Download the ZIP file")
            print("4. Extract the *_Manual-xccdf.xml file")
            print(f"5. Run: python fetch_stig.py --xccdf path/to/xccdf.xml -o {args.output}")
            sys.exit(1)
    else:
        parser.print_help()
        print("\n\nRecommended workflow:")
        print("1. Download Windows 11 STIG from https://public.cyber.mil/stigs/downloads/")
        print("2. Extract the ZIP and find the *_Manual-xccdf.xml file")
        print("3. Run: python fetch_stig.py --xccdf <path-to-xccdf.xml> -o stig.json")


if __name__ == '__main__':
    main()
